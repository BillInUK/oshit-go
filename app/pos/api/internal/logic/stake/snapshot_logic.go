package stake

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/app/pos/api/internal/task"
	"oshit-go/app/pos/api/types"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"time"
)

type StakeSnapShotLogic struct {
	prefix            string
	decimals          uint8
	ctx               context.Context
	srvCtx            *svc.ServiceContext
	db                *gorm.DB
	rd                redis.UniversalClient
	rs                redsync.Redsync
	baseClient        *posrpc.BaseClient
	rpcClient         *rpc.Client
	LightHouseAddress solana.PublicKey

	// 业务相关参数
	distLevel     int32                               // 向上递归奖励的层级
	serviceConfig *model.StakeRewardConfig            // 下发奖励配置
	starWhiteList map[string]model.StakeStarWhitelist // 星级用户白名单
	starLevelRule map[int32]model.StakeStarLevelRule  // 星级评定规则
	fixConfig     map[int32]model.StakeFixRateConfig  // 每日固定利息配置
	inviteRate    map[int32]model.StakeInviteRate
}

func NewStakeSnapShotLogic(ctx context.Context, srvCtx *svc.ServiceContext) *StakeSnapShotLogic {
	return &StakeSnapShotLogic{
		prefix:     "Stake快照业务 -",
		ctx:        ctx,
		srvCtx:     srvCtx,
		db:         srvCtx.DB,
		rd:         srvCtx.Redis,
		rs:         srvCtx.RedSync,
		rpcClient:  srvCtx.RpcClient,
		baseClient: srvCtx.BaseClient,

		fixConfig:         srvCtx.StakeFixConfig,
		distLevel:         srvCtx.StakeDistLevel,
		inviteRate:        srvCtx.StakeInviteRate,
		starWhiteList:     srvCtx.StakeStarWhitelist,
		starLevelRule:     srvCtx.StakeStarLevelRule,
		decimals:          uint8(srvCtx.TokenConfig.Decimals),
		LightHouseAddress: srvCtx.LightHouseAddress,
	}
}

// TakeStakeSnapShot 手动开启Stake快照
func (l *StakeSnapShotLogic) TakeStakeSnapShot() error {
	//if l.srvCtx.SystemConfig.Env == 0 {
	//	return response.FailWithMsg(c, "can not take snap shot on mainnet")
	//}
	taskCtx := &task.TaskContext{
		CoreContext:  l.srvCtx.CoreContext,
		RewardConfig: l.srvCtx.StakeRewardConfig,
	}
	snapShotTask := task.NewStakeSnapShotTask(taskCtx)
	snapShotTask.StartTaskManually()
	return nil
}

// ResetStakeSnapShot 重置Stake快照
func (l *StakeSnapShotLogic) ResetStakeSnapShot() error {
	//if l.srvCtx.SystemConfig.Env == 0 {
	//	return response.FailWithMsg(c, "can not take snap shot on mainnet")
	//}

	// 使用 Exec 执行多条 SQL 语句
	err := l.db.Exec(`
		delete from t_stake_snap_shot;
		delete from t_stake_reward;
		delete from t_stake_reward_claim;
	`).Error
	if err != nil {
		return err
	}
	return nil
}

// ProcessSnapShot 处理快照,发放普通用户奖励以及星级用户的极差奖励
func (l *StakeSnapShotLogic) ProcessSnapShot(msg entity.KafkaNewSnapShotMsg) {
	snapShotDay := msg.MsgContent
	maxDepth := 20

	// 计算每日固定利息
	rewardFixMap, snapShotMap, err := l.rewardOrdinaryStaker(snapShotDay)
	if err != nil {
		log.Errorf("%s 发放普通质押用户每日固定利息错误: %v", l.prefix, err)
		return
	}
	// 基于每日固定利息计算邀请奖励
	rewardInviteMap, err := l.rewardInviter(rewardFixMap)
	if err != nil {
		log.Errorf("%s 发放普通质押用户邀请奖励错误: %v", l.prefix, err)
		return
	}
	// 基于每日固定利息和快照设置星级
	starredMap := l.setStakeStarLevel(rewardInviteMap, snapShotMap, snapShotDay)
	// 计算星级奖励
	l.rewardGroup(starredMap, snapShotMap, snapShotDay, maxDepth)
	// 过期昨日的奖励
	l.expireRewards(snapShotDay)
}

// rewardOrdinaryStaker 发放普通质押用户的奖励以及其邀请人的奖励
// 经过此函数
// StarLevel: 0
// Base: 质押金额的Base
// RewardAmount: 每日固定利息
// Rate: 每日固定利息费率
func (l *StakeSnapShotLogic) rewardOrdinaryStaker(snapShotDay time.Time) (map[string]model.StakeReward, map[string]types.StakeSnapShotDetail, error) {
	var err error
	// Step 1: 使用 JOIN 查询 t_sol_stake_snap_shot 和 t_sol_stake_star_level_config 表符合条件的数据，限定时间段
	var snapShots []model.StakeSnapShot
	fixConfigMap := l.fixConfig

	log.Infof("%s 发放普通质押用户奖励 - 快照日期: %v\n", l.prefix, snapShotDay)

	db := l.db
	if err = db.Table(model.TableNameStakeSnapShot).Where("snap_day = ?", snapShotDay).Scan(&snapShots).Error; err != nil {
		log.Errorf("%s 根据日期查询快照失败: %v", l.prefix, err)
		return nil, nil, err
	}
	log.Infof("%s 找到快照信息: %v", l.prefix, snapShots)

	// Step 2: 创建批量插入的数据，根据质押类型的不同发放不同类型的奖励
	var rewardRecords []model.StakeReward
	rewardRecordMap := make(map[string]model.StakeReward)
	snapShotMap := make(map[string]types.StakeSnapShotDetail)

	// Step 3: 根据快照计算出来每个地址需要奖励的金额
	for _, snapShot := range snapShots {
		if snapShot.Amount == 0.0 || snapShot.NativeAccount == "" {
			continue
		}
		rateConfig, exist := fixConfigMap[snapShot.StakeType]
		if !exist {
			log.Errorf("%s 发放质押奖励错误，地址 %s 的质押类型 %d 不支持", l.prefix, snapShot.NativeAccount, snapShot.StakeType)
			continue
		}
		snapShotBase := snapShot.Amount * rateConfig.FixRate / 100
		rewardAmount := snapShotBase / 365

		// 统计快照里里面的总质押量和质押Base
		if record, exist := snapShotMap[snapShot.NativeAccount]; !exist {
			record := types.StakeSnapShotDetail{
				NativeAccount: snapShot.NativeAccount,
				SnapTotal:     snapShot.Amount,
				SnapBase:      snapShotBase,
			}
			snapShotMap[snapShot.NativeAccount] = record
		} else {
			record.SnapTotal += snapShot.Amount
			record.SnapBase += snapShotBase
			snapShotMap[snapShot.NativeAccount] = record
		}

		// 将奖励信息写入到奖励发放Map
		if record, exist := rewardRecordMap[snapShot.NativeAccount]; !exist {
			record := model.StakeReward{
				NativeAccount: snapShot.NativeAccount,
				GroupID:       snapShot.NativeAccount,
				StarLevel:     0,
				Base:          snapShotBase,
				Rate:          rateConfig.FixRate,
				RewardAmount:  rewardAmount,
				RewardType:    types.StakeFixed,
				RewardState:   int32(constants.RewardStateInit),
				Pending:       false,
				Starred:       false,
				SnapDay:       snapShotDay,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			rewardRecordMap[snapShot.NativeAccount] = record
		} else {
			record.Base += snapShotBase
			record.RewardAmount += rewardAmount
			rewardRecordMap[snapShot.NativeAccount] = record
		}
	}

	// Step 4: 遍历map，然后将每日固定利息奖励放到数组里面
	for _, record := range rewardRecordMap {
		// 插入数据库前将base设置为snapShotAmount里面的Base
		rewardRecords = append(rewardRecords, record)
	}

	// Step5: 创建邀请人奖励记录
	if err = db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "native_account"},
				{Name: "snap_day"},
				{Name: "reward_type"},
				{Name: "starred"},
			},
			DoNothing: true, // 如果冲突则跳过
		}).CreateInBatches(rewardRecords, 1000).Error; err != nil {
		log.Errorf("failed to insert snapShots in batch: %v", err)
		return nil, nil, err
	}

	// 返回每个用户的每日固定利息
	return rewardRecordMap, snapShotMap, nil
}

// rewardInviter 发放邀请人奖励
// StarLevel: 0
// Base: 被设置为每日固定利息
// Rate: 费率
// RewardAmount: 经过本函数将被设置为邀请人奖励
func (l *StakeSnapShotLogic) rewardInviter(rewardMap map[string]model.StakeReward) (map[string]model.StakeReward, error) {
	var rewardRecords []model.StakeReward

	// 清空所有的奖励金额，并将Base设置为每日固定利息
	for inviter, _ := range rewardMap {
		inviterRecord := rewardMap[inviter]
		inviterRecord.Base = inviterRecord.RewardAmount
		inviterRecord.RewardType = types.StakeInvite
		inviterRecord.RewardAmount = 0
		rewardMap[inviter] = inviterRecord
	}

	// 遍历map发放邀请人奖励
	for invitee, record := range rewardMap {
		// 把rewardMap里面的地址当成是被邀请人，逐个计算奖励上级邀请人的信息
		inviteRewardMap, err := l.distributeInviteRewardsWithRecursiveCTE(l.db, rewardMap, invitee, record.SnapDay, record.Base)
		if err != nil {
			log.Errorf("%s 创建地址 %s 邀请人奖励记录错误: %v", l.prefix, record.NativeAccount, err)
			return rewardMap, err
		}

		// 累加奖励给到邀请人
		for inviter, amount := range inviteRewardMap {
			if inviter == "7e8jvFAsiwbYRNzttL2WeAvyzHF7HFzUhaJWQAWJ1G9A" {
				log.Infof("%s 奖励邀请人 7e8jvFAsiwbYRNzttL2WeAvyzHF7HFzUhaJWQAWJ1G9A 奖励金额 %v", l.prefix, amount)
			}
			inviterRecord := rewardMap[inviter]
			inviterRecord.RewardType = types.StakeInvite
			inviterRecord.RewardAmount += amount
			rewardMap[inviter] = inviterRecord
		}
	}

	// 将邀请项目组织成数组
	for _, record := range rewardMap {
		rewardRecords = append(rewardRecords, record)
	}

	// 批量插入邀请人奖励数据
	if err := l.db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "native_account"},
				{Name: "snap_day"},
				{Name: "reward_type"},
				{Name: "starred"},
			},
			DoNothing: true, // 如果冲突则跳过
		}).CreateInBatches(rewardRecords, 1000).Error; err != nil {
		log.Errorf("%s 批量插入每日固定利息奖励记录失败: %v", l.prefix, err)
		return rewardMap, err
	}

	return rewardMap, nil
}

// setStakeStarLevel 根据用户质押金额和快照信息计算并设置星级等级
// 功能说明：
//   - 遍历所有用户质押奖励记录，根据质押总额从配置规则中获取对应的星级
//   - 设置每个用户的星级、费率，并记录计算日志
//
// 参数:
//   - rewardMap: 用户质押奖励映射，key为原生账户地址，value为质押奖励信息
//   - snapShotMap: 质押快照详情映射，key为原生账户地址，value为快照详情
//   - snapShotDay: 快照日期，用于星级计算的时间基准
//
// 返回值:
//   - map[string]model.StakeReward: 更新后的用户质押奖励映射，包含星级和费率信息
//
// 处理逻辑:
//   - 从全局配置获取星级规则
//   - 第一轮循环：计算每个账户的星级并更新费率
//   - 第二轮循环：记录每个账户的星级分配结果
//
// 字段说明:
//   - StarLevel: 用户星级等级 (1-5星等)
//   - Rate: 对应星级的费率系数
//   - Base/RewardAmount: 每日固定利息基础值
func (l *StakeSnapShotLogic) setStakeStarLevel(rewardMap map[string]model.StakeReward, snapShotMap map[string]types.StakeSnapShotDetail, snapShotDay time.Time) map[string]model.StakeReward {
	rateConfig := l.starLevelRule
	// 给每个人计算星级
	for _, record := range rewardMap {
		stakeTotal := snapShotMap[record.NativeAccount].SnapTotal
		stakeStarLevel, _, _ := l.GetStakeStarLevelFromConfig(record.NativeAccount, stakeTotal, snapShotDay)
		record.StarLevel = stakeStarLevel
		record.Rate = rateConfig[stakeStarLevel].Rate
		rewardMap[record.NativeAccount] = record
	}
	for _, record := range rewardMap {
		log.Infof("%s - 地址 %s 质押星级 %d", l.prefix, record.NativeAccount, record.StarLevel)
	}
	return rewardMap
}

// rewardGroup 发放邀请关系当中的星级用户奖励
func (l *StakeSnapShotLogic) rewardGroup(rewardMap map[string]model.StakeReward, snapShotMap map[string]types.StakeSnapShotDetail, snapShotDay time.Time, maxDepth int) {
	var err error
	var inviteChains []types.InviteNode

	log.Infof("%s 发放质押激励奖励团队奖励部分 - 快照日期: %v\n", l.prefix, snapShotDay)

	for _, record := range rewardMap {
		log.Infof("%s 发放质押激励奖励团队奖励部分 地址: %s 星级 %d Base %v 费率 %v",
			l.prefix, record.NativeAccount, record.StarLevel, record.Base, record.Rate)
	}

	// 查询极差奖励数据
	err = l.db.Raw(`
		with recursive invite_chain as (
			-- 初始查询，从没有下级邀请人的节点开始
			select
				r.inviter,
				r.invitee,
				r.inviter_level,
				r.invitee as group_id,
				1 as depth,
				r.invitee || '-' || r.inviter as path -- 用来追踪路径，避免循环
			from
				t_invite_relation r
			where
				not exists (
					select 1
					from t_invite_relation sub_r
					where sub_r.inviter = r.invitee
				)
			union all
			-- 递归查询，从当前节点的 Inviter 向上追溯
			select
				r.inviter,
				r.invitee,
				r.inviter_level,
				ic.group_id,
				ic.depth + 1 as depth,
				ic.path || '-' || r.inviter  -- 添加路径信息
			from
				t_invite_relation r
					join
				invite_chain ic
				on
					ic.inviter = r.invitee
			where
				not ic.path like '%' || r.inviter || '%' -- 排除循环路径
		)
		select
			ic.group_id,
			ic.inviter,
			ic.invitee,
			-- 如果 inviter 为空，说明是顶级节点，level - 1
			case
				when ic.inviter is null then ic.inviter_level - 1
			else ic.inviter_level
			end as level
		from
			invite_chain ic
				inner join
			t_stake_snap_shot ss
			on
				ss.native_account = ic.inviter
		where
			ss.snap_day = cast(? as date)
		group by
			ic.group_id, ic.inviter, ic.invitee, ic.inviter_level
		order by
			ic.group_id, ic.inviter_level, ic.invitee, ic.inviter;
	
	`, snapShotDay).Scan(&inviteChains).Error
	if err != nil {
		log.Errorf("%s 发放邀请关系内的星级用户奖励 - 查询错误: %v", l.prefix, err)
		return
	}

	// 0.给每个节点设置星级，费率，每日固定利息
	for index, node := range inviteChains {
		if node.Inviter == "" {
			continue
		}
		// 因为Base是每日固定利息，但是实际上我们计算极差奖励的时候，是按照质押金额的base来进行计算的
		starLevel := rewardMap[node.Inviter].StarLevel
		starRate := rewardMap[node.Inviter].Rate
		amount := snapShotMap[node.Inviter].SnapTotal
		base := snapShotMap[node.Inviter].SnapBase
		inviteChains[index].StarLevel = int(starLevel)
		inviteChains[index].Rate = starRate
		inviteChains[index].Base = base
		inviteChains[index].Amount = amount
		log.Infof("%s 计算团队奖励 - groupId %s 质押节点邀请人 %s 质押节点被邀请人 %s 星级 %d Base %v 费率 %v",
			l.prefix,
			inviteChains[index].GroupId,
			inviteChains[index].Inviter,
			inviteChains[index].Inviter,
			inviteChains[index].StarLevel,
			inviteChains[index].Base,
			inviteChains[index].Rate)
	}

	// 1. 按 GroupID 分组，生成邀请树分支数组
	groupedBranch := make(map[string][]types.InviteNode)
	for _, inviteNode := range inviteChains {
		groupedBranch[inviteNode.GroupId] = append(groupedBranch[inviteNode.GroupId], inviteNode)
	}

	// 2. 遍历每个邀请分支，然后在每个邀请分支的末尾加上级别最低被邀请的质押用户
	for groupId, branch := range groupedBranch {
		var maxInviteLevelNode types.InviteNode
		for _, inviteNode := range branch {
			if inviteNode.Level > maxInviteLevelNode.Level {
				maxInviteLevelNode = inviteNode
			}
		}
		log.Infof("%s groupId %s 最大邀请级 邀请人 %s 被邀请人 %s 邀请级别 %d", l.prefix, groupId, maxInviteLevelNode.Inviter, maxInviteLevelNode.Invitee, maxInviteLevelNode.Level)

		// 3. 将查询结果添加到邀请树的每个分支的后面，因为上面的递归查询没有包含最后被邀请的节点
		//inviterNativeAccount := maxInviteLevelNode.Invitee
		inviteeNativeAccount := maxInviteLevelNode.Invitee
		base := snapShotMap[inviteeNativeAccount].SnapBase
		starLevel := int(rewardMap[inviteeNativeAccount].StarLevel)
		rate := rewardMap[inviteeNativeAccount].Rate
		inviteLevel := maxInviteLevelNode.Level + 1
		lastInviteNode := types.InviteNode{
			GroupId:   groupId,
			Inviter:   inviteeNativeAccount,
			Invitee:   inviteeNativeAccount,
			Level:     inviteLevel,
			Base:      base,
			StarLevel: starLevel,
			Rate:      rate,
		}
		groupedBranch[groupId] = append(groupedBranch[groupId], lastInviteNode)
	}

	// 遍历每条邀请分支，然后依次计算出每个节点的极差奖励的Base
	selfFlagMap := make(map[string]bool)
	topFlagMap := make(map[string]bool)
	totalRewardMap := make(map[string]model.StakeReward)
	for groupId, chain := range groupedBranch {
		// 自下而上计算每个节点极差奖励的base
		for i := 0; i < len(chain); i++ {
			if chain[i].Inviter == "" {
				continue
			}
			l.calculateStakeBase(chain, i, maxDepth, selfFlagMap, topFlagMap)
		}
		for _, node := range chain {
			if node.Inviter == "" {
				continue
			}
			rewardItem, exist := totalRewardMap[node.Inviter]
			if !exist && node.StarLevel > 0 && node.Base > 0 {
				rewardRecord := model.StakeReward{
					GroupID:       groupId,
					NativeAccount: node.Inviter,
					StarLevel:     int32(node.StarLevel),
					Base:          node.Base,
					Rate:          node.Rate,
					RewardAmount:  node.Base / 365,
					RewardType:    types.StakeStarGroup,
					RewardState:   int32(constants.RewardStateInit),
					Pending:       false,
					Starred:       node.StarLevel > 0,
					SnapDay:       snapShotDay,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				totalRewardMap[node.Inviter] = rewardRecord
			} else {
				if node.Base > 0 {
					rewardItem.Base += node.Base
					rewardItem.RewardAmount += node.Base / 365
					totalRewardMap[node.Inviter] = rewardItem
				}
			}
		}
	}

	table := l.db.Table(model.TableNameStakeReward)
	for _, rewardItem := range totalRewardMap {
		if err = table.Clauses(
			clause.OnConflict{
				Columns: []clause.Column{
					{Name: "native_account"},
					{Name: "snap_day"},
					{Name: "reward_type"},
					{Name: "starred"},
				},
				DoNothing: true, // 如果冲突则跳过
			}).Create(&rewardItem).Error; err != nil {
			log.Errorf("%s 新增邀请关系内星级用户固定奖励 - 插入奖励记录错误: %v", l.prefix, err)
			break
		}
	}
}

func (l *StakeSnapShotLogic) distributeInviteRewardsWithRecursiveCTE(tx *gorm.DB, rewardMap map[string]model.StakeReward, invitee string, day time.Time, totalFix float64) (map[string]float64, error) {
	// 获取所有级别的费率
	//var rewardRecords []model.StakeReward
	distLevel := l.distLevel
	inviteRates := l.inviteRate

	inviteRewardMap := make(map[string]float64)
	// 使用递归CTE查找有效的邀请链
	var inviteChain []struct {
		InviterNativeAccount string  `gorm:"column:inviter"`
		Level                int32   `gorm:"column:level"`
		RewardAmount         float64 `gorm:"column:reward_amount"`
	}

	recursiveQuery := `
        with recursive invite_chain AS (
            -- 基础情况：直接邀请人
            select 
                ir.inviter,
                1 as level
            from public.t_invite_relation ir
            where ir.invitee = ?
            
            union all
            
            -- 递归情况：间接邀请人
            select 
                ir.inviter,
                ic.level + 1 AS level
            from public.t_invite_relation ir
            inner join invite_chain ic on ir.invitee = ic.inviter
            where ic.level < ?
        )
        select 
            ic.inviter,
            ic.level AS level
        from invite_chain ic
        inner join public.t_stake_snap_shot ss on ic.inviter = ss.native_account 
            and ss.snap_day = ?
        order by ic.level
    `

	if err := tx.Raw(recursiveQuery, invitee, distLevel, day).Scan(&inviteChain).Error; err != nil {
		log.Errorf("%s 递归查询邀请链错误: %v", l.prefix, err)
		return inviteRewardMap, fmt.Errorf("递归查询邀请链失败: %v", err)
	}
	if len(inviteChain) == 0 {
		log.Infof("%s 未找到有效的邀请链 for account: %s", l.prefix, invitee)
		return inviteRewardMap, nil
	}

	// 为邀请链中的每个有效邀请人发放奖励
	for _, inviter := range inviteChain {
		// 查看邀请人是否在质押快照信息里面
		rate := inviteRates[inviter.Level].Rate
		if _, exist := rewardMap[inviter.InviterNativeAccount]; !exist {
			continue
		}
		// 如果存在，则将奖励累加到RewardAmount里面
		rewardAmount := totalFix * rate / 100
		inviterRewardAmount := inviteRewardMap[inviter.InviterNativeAccount]
		inviterRewardAmount += rewardAmount
		inviteRewardMap[inviter.InviterNativeAccount] = inviterRewardAmount
		log.Infof("%s 成功为被邀请人 %s 的邀请人 %s 分发了 %d 级邀请奖励 奖励金额 %v", l.prefix, invitee, inviter.InviterNativeAccount, inviter.Level, inviter.RewardAmount)
	}

	return inviteRewardMap, nil
}

// expireRewards 快照完成后将过去的奖励设置为超时
func (l *StakeSnapShotLogic) expireRewards(snapShotDay time.Time) {
	var err error
	table := l.db.Table(model.TableNameStakeReward)
	if err = table.Where("snap_day < cast(? as date) and reward_type in (?,?,?,?)",
		snapShotDay,
		types.StakeInvite,
		types.StakeStarIndividual,
		types.StakeStarIndividual,
		types.StakeStarGroup,
	).Update("reward_state", constants.RewardStateFailed).Error; err != nil {
		log.Errorf("%s 设置奖励为过期失败错误:%v", l.prefix, err)
	}
}

func (l *StakeSnapShotLogic) getGroupStakeAmount(rootAccount string, snapShotDay time.Time) (float64, error) {
	var totalHoldAmount float64
	query := `
		with recursive invite_tree as (
			select
				inviter as account
			from
				public.t_invite_relation
			where
				inviter = ?  -- 输入邀请人的账户
			union
			select
				r.invitee
			from
				public.t_invite_relation r
			inner join
				invite_tree it
			on
				r.inviter = it.account
		),
		group_accounts as (
			select distinct(account)
			from invite_tree
			where account != ?  -- 排除掉邀请人自己的账户
		),
		filtered_snap_shot as (
			select *
			from t_stake_snap_shot
			where snap_day = cast(? as date)
		)
		select
			coalesce(sum(fss.amount), 0) as total_hold_amount
		from
			group_accounts gp
		left join
			filtered_snap_shot fss
		on
			gp.account = fss.native_account;
	`
	err := l.db.Raw(query, rootAccount, rootAccount, snapShotDay).Scan(&totalHoldAmount).Error
	if err != nil {
		return 0, err
	}

	return totalHoldAmount, nil
}

// GetStakeStarLevelFromConfig 评定质押星级
func (l *StakeSnapShotLogic) GetStakeStarLevelFromConfig(owner string, amount float64, snapShotDay time.Time) (int32, float64, error) {
	config, exist := l.starWhiteList[owner]
	// 如果存在则以数据库配置的为主
	if exist {
		return config.StarLevel, config.Rate, nil
	} else {
		// 如果不存在则根据持币数量和下级数量来判断是否是星级用户
		var stakeLevel, groupLevel int32 = 0, 0
		var stakeLevelRate, groupLevelRate = 0.0, 0.0
		for _, rule := range l.starLevelRule {
			if uint64(amount) >= uint64(rule.Amount) {
				stakeLevel = rule.StarLevel
				stakeLevelRate = rule.Rate
				log.Infof("%s 地址 %s 根据质押量 %.3f  规则质押量 %.3f 评定星级 %d", l.prefix, owner, amount, rule.Amount, stakeLevel)
				continue
			}
			log.Infof("%s 地址 %s 根据质押量 %.3f  规则质押量 %.3f 最终评定星级 %d", l.prefix, owner, amount, rule.Amount, stakeLevel)
			break
		}

		// 如果质押量不够星级则直接返回星级为0
		if stakeLevel == 0 {
			return 0, 0.0, nil
		}
		// 质押量达到星级后，再判断团队持币量
		groupStakeAmount, err := l.getGroupStakeAmount(owner, snapShotDay)
		if err != nil {
			return 0, 0.0, nil
		}
		if stakeLevel < 5 {
			// 如果持币量的星级小于5星，则判断团队持币星级的时候包含账户自己的持币量
			groupStakeAmount += amount
			for _, rule := range l.starLevelRule {
				if uint64(groupStakeAmount) >= uint64(rule.GroupAmount) {
					groupLevel = rule.StarLevel
					groupLevelRate = rule.Rate
					log.Infof("%s 地址 %s 根据团队质押量 %.3f 规则邀请质押量 %.3f 评定星级 %d", l.prefix, owner, groupStakeAmount, rule.GroupAmount, groupLevel)
					continue
				}
				log.Infof("%s 地址 %s 根据团队质押量 %.3f 规则邀请质押量 %.3f 最终评定星级 %d", l.prefix, owner, groupStakeAmount, rule.GroupAmount, groupLevel)
				break
			}
		} else {
			// 则先判断下级持币量是否达到了5星，如果达到5星则直接返回5星
			if groupStakeAmount >= l.starLevelRule[4].GroupAmount {
				groupLevel = l.starLevelRule[4].StarLevel
				groupLevelRate = l.starLevelRule[4].Rate
				return groupLevel, groupLevelRate, nil
			} else {
				// 如果下级持币量没有达到5星，则判断总持币量是否能达到最高4星
				groupStakeAmount += amount
				for _, rule := range l.starLevelRule {
					if groupStakeAmount >= rule.GroupAmount {
						groupLevel = rule.StarLevel
						groupLevelRate = rule.Rate
						continue
					}
					break
				}
				// 星级最多为4星
				groupLevel = min(groupLevel, 4)
				groupLevelRate = min(groupLevelRate, l.starLevelRule[3].Rate)
				return groupLevel, groupLevelRate, nil
			}
		}
		return min(stakeLevel, groupLevel), min(stakeLevelRate, groupLevelRate), nil
	}
}

func (l *StakeSnapShotLogic) calculateStakeBase(records []types.InviteNode, index int, maxOffset int, selfFlagMap map[string]bool, topFlagMap map[string]bool) {
	// 如果分支的长度只有1则Base是当前的 每个固定利息 * 星级费率
	if len(records) == 1 {
		selfFlagMap[records[index].Inviter] = true
		records[index].Base = records[index].Base * records[index].Rate / 100
		return
	}
	// 获取当前记录
	record := records[index]

	// 如果当前记录的 StarLevel 为 0，则直接将 Base 设置为 0，并结束计算
	if record.StarLevel == 0 {
		records[index].Base = 0
		selfFlagMap[record.Inviter] = true
		return
	}

	// 判断是否要相加当前的base
	base := 0.0
	currentRate := record.Rate
	selfBase := record.Base * record.Rate / 100
	_, exist := selfFlagMap[record.Inviter]
	if !exist {
		base += selfBase
		selfFlagMap[record.Inviter] = true
	}

	// 开始往后计算，最多计算n个元素（这里n = 20）
	firstRun := true
	deltaRate := 0.0
	maxNextRate := currentRate
	for i := index + 1; i < len(records) && i < index+maxOffset; i++ {
		lastRecord := records[i-1]
		nextRecord := records[i]

		// 如果Rate[i] >= Rate[i-1]，结束计算
		if nextRecord.Rate >= currentRate {
			break
		}
		if nextRecord.Rate < lastRecord.Rate {
			// 如果Rate[i] < Rate[i-1]，按照之前的极差计算
			// 如果是首次运行则改变deltaRate
			if firstRun {
				deltaRate = currentRate - nextRecord.Rate
				maxNextRate = nextRecord.Rate
				firstRun = false
			}
			// 如果当前节点已经向下级收过极差，则不在收取
			if _, exist = topFlagMap[record.Inviter+"-"+nextRecord.Inviter]; exist {
				continue
			} else {
				topFlagMap[record.Inviter+"-"+nextRecord.Inviter] = true
				base += nextRecord.Base * deltaRate / 100
			}
		} else {
			// 如果Rate[i] > Rate[i-1]，但是小于当前节点的rate，则按新极差进行计算
			if nextRecord.Rate > maxNextRate && nextRecord.Rate < currentRate {
				deltaRate = currentRate - nextRecord.Rate
				maxNextRate = nextRecord.Rate
			}
			// 如果当前节点已经向下级收过极差，则不在收取
			if _, exist = topFlagMap[record.Inviter+"-"+nextRecord.Inviter]; exist {
				continue
			} else {
				topFlagMap[record.Inviter+"-"+nextRecord.Inviter] = true
				base += nextRecord.Base * deltaRate / 100
			}
		}
	}

	// 更新记录的Base值
	records[index].Base = base
	return
}
