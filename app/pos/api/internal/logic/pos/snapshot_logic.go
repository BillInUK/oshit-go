package pos

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/app/pos/api/internal/task"
	"oshit-go/app/pos/api/types"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"strings"
	"time"
)

type PosSnapShotLogic struct {
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
	serviceConfig *model.PosRewardConfig
	starLevelRule map[int32]model.PosStarLevelRule
	whiteList     map[string]model.PosStarWhitelist
}

func NewPosSnapShotLogic(ctx context.Context, srvCtx *svc.ServiceContext) *PosSnapShotLogic {
	return &PosSnapShotLogic{
		prefix:            "Pos快照业务 -",
		ctx:               ctx,
		srvCtx:            srvCtx,
		db:                srvCtx.DB,
		rd:                srvCtx.Redis,
		rs:                srvCtx.RedSync,
		rpcClient:         srvCtx.RpcClient,
		baseClient:        srvCtx.BaseClient,
		decimals:          uint8(srvCtx.TokenConfig.Decimals),
		LightHouseAddress: srvCtx.LightHouseAddress,

		serviceConfig: srvCtx.PosRewardConfig,
		starLevelRule: srvCtx.PosStarLevelRule,
		whiteList:     srvCtx.PosWhiteListMap,
	}
}

// TakeStakeSnapShot 手动快照
func (l *PosSnapShotLogic) TakeStakeSnapShot() error {
	taskCtx := &task.TaskContext{
		CoreContext:  l.srvCtx.CoreContext,
		RewardConfig: l.srvCtx.StakeRewardConfig,
	}
	snapShotTask := task.NewPosSnapShotTask(taskCtx)
	snapShotTask.StartTaskManually()
	return nil
}

// ResetStakeSnapShot 重置快照
func (l *PosSnapShotLogic) ResetStakeSnapShot() error {
	// 使用 Exec 执行多条 SQL 语句
	err := l.db.Exec(`
		delete from t_pos_snap_shot;
		delete from t_pos_reward;
		delete from t_pos_reward_claim;
	`).Error
	if err != nil {
		return err
	}
	return nil
}

// ProcessSnapShot 处理快照,发放普通用户奖励以及星级用户的极差奖励
func (l *PosSnapShotLogic) ProcessSnapShot(msg entity.KafkaNewSnapShotMsg) {
	snapShotDay := msg.MsgContent
	maxDepth := 20

	l.reConstructInviteRelation()
	l.rewardOrdinaryUser(snapShotDay)
	l.rewardStaredUserOrphan(snapShotDay)
	l.rewardStaredUserDeterMineInvite(snapShotDay, maxDepth)
	l.expireLastDayPosReward(snapShotDay)
}

// reConstructInviteRelation 重新建立邀请层级防止因为层级数错误导致奖励发放异常
func (l *PosSnapShotLogic) reConstructInviteRelation() {
	query := `
		-- 创建一个递归查询，计算层级关系
		with recursive invite_hierarchy as (
		  -- 基层的邀请人，默认 DistLevel 为 1
		  select
			record_id,
			inviter,
			invitee,
			1 as level
		  from
			t_invite_relation
		  where
			inviter not in (select invitee from t_invite_relation)

		  union all
			
		  -- 递归查找下一级邀请关系
		  select
			r.record_id,
			r.inviter,
			r.invitee,
			ih.level + 1 as level
		  from
			t_invite_relation r
			  inner join
			invite_hierarchy ih
			on
			  r.inviter = ih.invitee
		)
		-- 更新原表中的 DistLevel 字段
		update t_invite_relation as t
		set level = ih.level
		  from invite_hierarchy ih
		where t.record_id = ih.record_id;
	`
	if err := l.db.Exec(query).Error; err != nil {
		log.Fatalf("pos业务 - 重新确定邀请等级错误 %v", err)
	}
}

// expireLastDayPosReward 快照完成后将过去的奖励设置为超时
func (l *PosSnapShotLogic) expireLastDayPosReward(snapShotDay time.Time) {
	var err error
	table := l.db.Table(model.TableNamePosReward)
	if err = table.Where("snap_day < cast(? as date)", snapShotDay).Update("reward_state", -1).Error; err != nil {
		log.Errorf("pos业务 - 设置奖励为过期失败错误:%v", err)
	}
}

// rewardOrdinaryUser 发送普通用户奖励
func (l *PosSnapShotLogic) rewardOrdinaryUser(snapShotDay time.Time) {
	var err error
	// Step 1: 使用 JOIN 查询 t_pos_snap_shot 和 t_sol_pos_star_level_config 表符合条件的数据，限定时间段
	var rewards []model.PosReward
	var missionConfig model.PosMissionConfig

	log.Infof("pos业务 - 发放普通用户奖励 - 快照日期: %v\n", snapShotDay)

	db := l.db
	if err = db.Table(model.TableNamePosMissionConfig).Where("reward_type = ? and starred = ?", types.PosFixedIncome, false).First(&missionConfig).Error; err != nil {
		log.Errorf("pos业务 - 发放普通用户固定收益，无法找到固定收益的费率配置，错误: %v", err)
		return
	}
	if err = db.Raw(`select native_account,amount as base from public.t_pos_snap_shot ss where snap_day=cast(? as date)`, snapShotDay).Scan(&rewards).Error; err != nil {
		log.Errorf("failed to perform join query with date range: %v", err)
		return
	}
	// Step 2: 创建批量插入的数据
	now := time.Now()
	rate := missionConfig.Rate / 100
	rewardRecords := make([]model.PosReward, len(rewards))
	for i, reward := range rewards {
		if reward.Base == 0.0 || reward.NativeAccount == "" {
			continue
		}
		rewardRecords[i] = model.PosReward{
			NativeAccount: reward.NativeAccount,
			StarLevel:     0,
			Base:          reward.Base,
			Rate:          missionConfig.Rate,
			RewardAmount:  reward.Base * rate / 365,
			RewardType:    types.PosFixedIncome,
			RewardState:   0,
			Pending:       false,
			Starred:       false,
			SnapDay:       snapShotDay,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
	}

	// Step 3: 批量插入到 t_sol_pos_reward 表
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
		log.Errorf("failed to insert rewards in batch: %v", err)
		return
	}

	return
}

func (l *PosSnapShotLogic) rewardStaredUserOrphan(snapShotDay time.Time) {
	var err error
	var records []types.InviteNode

	log.Infof("pos业务 - 发放孤点星级用户奖励 - 快照日期: %v\n", snapShotDay)

	var missionConfig model.PosMissionConfig
	err = l.db.Table(model.TableNamePosMissionConfig).
		Where("reward_type = ? and starred = ?", 0, true).First(&missionConfig).Error
	if err != nil {
		log.Errorf("pos业务 - 发放邀请关系外星级用户奖励，无法找到固定收益的费率配置，错误: %v", err)
		return
	}

	// 使用 Join 将 t_pos_snap_shot 和 t_sol_pos_star_level_config 表联合，获取符合条件且在指定时间段内的数据
	holdRate := missionConfig.Rate

	// 查询极差奖励数据
	err = l.db.Raw(`
		select
		  p.native_account as inviter,
		  NULL AS invitee,
		  1 AS level,
		  p.amount,
		  p.range_base,
		  p.native_account as group_id,
		  1 as depth
		from
		  public.t_pos_snap_shot p
		where
		  not exists (
			select 1
			from t_invite_relation r
			where r.inviter = p.native_account
			   or r.invitee = p.native_account
		  )
		  and p.star_level > 0 AND p.snap_day = cast(? as date)
		order by
		  group_id, level, inviter;
		`, snapShotDay).Scan(&records).Error
	if err != nil {
		// 错误处理
		log.Errorf("pos业务 - 发放邀请关系外星级用户奖励，查询错误: %v", err)
		return
	}

	// 遍历每条分支，然后计算出固定奖励，固定极差奖励
	for _, r := range records {
		table := l.db.Table(model.TableNamePosReward)
		if r.Base == 0.0 || r.Inviter == "" {
			continue
		}
		starLevel, starRate, _ := l.GetPosStarLevelFromConfig(r.Inviter, r.Amount, snapShotDay)
		rangeRewardRecord := model.PosReward{
			NativeAccount: r.Inviter,
			StarLevel:     starLevel,
			Base:          r.Amount * starRate / 100,
			Rate:          holdRate,
			RewardAmount:  r.Amount * starRate / 100 * holdRate / 100 / 365,
			RewardType:    types.PosFixedIncome,
			RewardState:   0,
			Pending:       false,
			Starred:       true,
			SnapDay:       snapShotDay,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err = table.Clauses(
			clause.OnConflict{
				Columns: []clause.Column{
					{Name: "native_account"},
					{Name: "snap_day"},
					{Name: "reward_type"},
					{Name: "starred"},
				},
				DoNothing: true, // 如果冲突则跳过
			}).Create(&rangeRewardRecord).Error; err != nil {
			log.Errorf("pos业务 - 发放邀请关系外星级用户奖励 - 插入奖励记录错误: %v", err)
			break
		}
	}
}

// 发放在邀请关系内的星级用户极差奖励
func (l *PosSnapShotLogic) rewardStaredUserDeterMineInvite(snapShotDay time.Time, maxDepth int) {
	var err error
	var inviteChains []types.InviteNode

	log.Infof("pos业务 - 发放邀请关系内的星级用户奖励 - 快照日期: %v\n", snapShotDay)

	var missionConfig model.PosMissionConfig
	err = l.db.Table(model.TableNamePosMissionConfig).
		Where("reward_type = ? AND starred = ?", 0, true).First(&missionConfig).Error
	if err != nil {
		log.Errorf("pos业务 - 发放邀请关系内的星级用户奖励，无法找到固定收益的费率配置，错误: %v", err)
		return
	}

	// 使用 Join 将 t_pos_snap_shot 和 t_sol_pos_star_level_config 表联合，获取符合条件且在指定时间段内的数据
	holdRate := missionConfig.Rate

	// 查询极差奖励数据
	err = l.db.Raw(`
			with recursive invite_chain AS (
				-- 初始查询，从没有下级邀请人的节点开始
				select
					r.inviter,
					r.invitee,
					r.level,
					r.invitee as group_id,
					1 as depth,
					r.invitee || '-' || r.inviter as path -- 用来追踪路径，避免循环
				from
					public.t_invite_relation r
				where
					not exists (
						select 1
						from public.t_invite_relation sub_r
						where sub_r.inviter = r.invitee
					)
				union all
				-- 递归查询，从当前节点的 Inviter 向上追溯
				select
					r.inviter,
					r.invitee,
					r.level,
					ic.group_id,
					ic.depth + 1 as depth,
					ic.path || '-' || r.inviter  -- 添加路径信息
				from
					public.t_invite_relation r
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
				-- 如果 Inviter 为空，说明是顶级节点，DistLevel - 1
				case
					when ic.inviter is null then ic.level - 1
					else ic.level
				end as level,
				ss.amount,
				ss.range_base
			from
				invite_chain ic
			inner join
				public.t_pos_snap_shot ss
			on
				ss.native_account= ic.inviter
			where
				ss.snap_day = cast(? as date)
			group by
				ic.group_id, ic.inviter, ic.invitee, ic.level,
				ss.amount, ss.range_base, ss.star_level, ss.rate
			order by
				ic.group_id, ic.level, ic.invitee, ic.inviter;
	`, snapShotDay).Scan(&inviteChains).Error
	if err != nil {
		log.Errorf("pos业务 - 发放邀请关系内的星级用户奖励 - 查询错误: %v", err)
		return
	}

	// 0.给每个节点设置星级
	for index, node := range inviteChains {
		if node.Inviter == "" {
			continue
		}
		starLevel, starRate, _ := l.GetPosStarLevelFromConfig(node.Inviter, node.Amount, snapShotDay)
		inviteChains[index].StarLevel = int(starLevel)
		inviteChains[index].Rate = starRate
	}

	// 1. 按 GroupId 分组
	grouped := make(map[string][]types.InviteNode)
	for _, chain := range inviteChains {
		grouped[chain.GroupId] = append(grouped[chain.GroupId], chain)
	}

	// 2. 找出每个 GroupId 中 DistLevel 最大的记录
	for groupId, chains := range grouped {
		var maxLevelChain types.InviteNode
		for _, chain := range chains {
			if chain.Level > maxLevelChain.Level {
				maxLevelChain = chain
			}
		}

		// 3. 使用最大 DistLevel 记录的 InviteeNativeAccount 查询 t_pos_snap_shot 表
		var posSnapshot model.PosSnapShot
		if err = l.db.Table(model.TableNamePosSnapShot).
			Where("native_account = ? and snap_day = cast(? as date)", maxLevelChain.Invitee, snapShotDay).
			First(&posSnapshot).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Infof("pos业务 - 发放邀请关系内的星级用户奖励 - 查询邀请关系最底层节点错误 %s: %v", maxLevelChain.Invitee, err)
		} else {
			// 4. 将查询结果添加到分组的最后面
			grouped[groupId] = append(grouped[groupId], types.InviteNode{
				GroupId:   groupId,
				Inviter:   maxLevelChain.Invitee,
				Invitee:   maxLevelChain.Invitee,
				Level:     maxLevelChain.Level,
				Amount:    posSnapshot.Amount,
				StarLevel: int(posSnapshot.StarLevel),
				Rate:      posSnapshot.Rate,
			})
		}
	}

	// 遍历每条分支，然后计算出固定奖励，固定极差奖励
	selfFlagMap := make(map[string]bool)
	topFlagMap := make(map[string]bool)
	totalRewardMap := make(map[string]model.PosReward)
	for groupId, chain := range grouped {
		for i := 0; i < len(chain); i++ {
			if chain[i].Inviter == "" {
				continue
			}
			l.CalculatePosBase(chain, i, maxDepth, selfFlagMap, topFlagMap)
		}
		for _, node := range chain {
			if node.Inviter == "" {
				continue
			}
			rewardItem, exist := totalRewardMap[node.Inviter]
			if !exist {
				rewardRecord := model.PosReward{
					GroupID:       groupId,
					NativeAccount: node.Inviter,
					StarLevel:     int32(node.StarLevel),
					Base:          node.Base,
					Rate:          holdRate,
					RewardAmount:  node.Base * holdRate / 100 / 365,
					RewardType:    types.PosFixedIncome,
					RewardState:   0,
					Pending:       false,
					Starred:       node.StarLevel > 0,
					SnapDay:       snapShotDay,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				totalRewardMap[node.Inviter] = rewardRecord
			} else {
				rewardItem.Base += node.Base
				rewardItem.RewardAmount += node.Base * holdRate / 100 / 365
				totalRewardMap[node.Inviter] = rewardItem
			}
		}
	}
	table := l.db.Table(model.TableNamePosReward)
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
			log.Errorf("pos业务 - 新增邀请关系内星级用户固定奖励 - 插入奖励记录错误: %v", err)
			break
		}
	}
}

func (l *PosSnapShotLogic) CalculatePosBase(records []types.InviteNode, index int, maxOffset int, selfFlagMap map[string]bool, topFlagMap map[string]bool) {
	// 如果分支的长度只有1则Base是当前的持币数量 * 星级费率
	if len(records) == 1 {
		selfFlagMap[records[index].Inviter] = true
		records[index].Base = records[index].Amount * records[index].Rate / 100
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
	selfBase := record.Amount * record.Rate / 100
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
				base += nextRecord.Amount * deltaRate / 100
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
				base += nextRecord.Amount * deltaRate / 100
			}
		}
	}

	// 更新记录的Base值
	records[index].Base = base
	return
}

func (l *PosSnapShotLogic) GetPosGroupHoldAmount(rootAccount string, snapShotDay time.Time) (float64, error) {
	var totalHoldAmount float64
	query := `
		with recursive invite_tree as (
			select
				inviter as account
			from
				t_invite_relation
			where
				inviter = ?  -- 输入邀请人的账户
			union
			select
				r.invitee
			from
				t_invite_relation r
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
			from t_pos_snap_shot
			where snap_day = cast(? as date)
		)
		from
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

func (l *PosSnapShotLogic) GetGroupTotalFixReward(rootAccount string, snapShotDay time.Time) (float64, error) {
	var totalFixReward float64
	query := `
		with recursive invite_tree as (
			select
				inviter as account
			from
				t_invite_relation
			where
				inviter = ?
			union

			select
				r.invitee
			from
				t_invite_relation r
			inner join
				invite_tree it
			on
				r.inviter = it.account
		),
		group_accounts as (
			select distinct(account)
			from invite_tree
		),filtered_rewards as (
			select * from t_sol_pos_reward where snap_day=cast(? as date) and reward_type=0
		)
		select
			coalesce(sum(fr.reward_amount), 0) as total_reward_amount
		from
			group_accounts gp
		left join
			filtered_rewards fr
		on
			gp.account = fr.native_account;
	`

	err := l.db.Raw(query, rootAccount, snapShotDay).Scan(&totalFixReward).Error
	if err != nil {
		return 0, err
	}

	return totalFixReward, nil
}

func (l *PosSnapShotLogic) QuerySnapShotByNativeAccount(currentAccount string) (*model.PosSnapShot, error) {
	var record model.PosSnapShot
	table := l.db.Table(model.TableNamePosSnapShot)
	if err := table.Where("native_account = ? ", currentAccount).Order("snap_day desc").First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Errorf("pos业务 - 根据地址 %s 日期 %v 查询快照信息错误: %v", currentAccount, time.Now(), err)
		return nil, err
	}
	return &record, nil
}

func (l *PosSnapShotLogic) QueryGroupInfoInSnapShot(currentAccount string) (*types.PosGroupInfo, error) {

	currentHoldAmount := 0.0
	latestSnapShotDay := time.Now()
	latestSnapShot, err := l.QuerySnapShotByNativeAccount(currentAccount)
	if err != nil {
		log.Errorf("pos业务 - 根据地址 %s 获取最近1天的快照错误: %v", currentAccount, err)
		return nil, err
	}
	if latestSnapShot != nil {
		currentHoldAmount = latestSnapShot.Amount
		latestSnapShotDay = latestSnapShot.SnapDay
	}
	inviterNativeAccount, err := l.GetInviterAccount(currentAccount)
	if err != nil {
		log.Errorf("pos业务 - 根据地址 %s 获取节点邀请人错误: %v", currentAccount, err)
		return nil, err
	}
	groupHoldAmount, err := l.GetPosGroupHoldAmount(currentAccount, latestSnapShotDay)
	if err != nil {
		log.Errorf("pos业务 - 根据地址 %s 获取团队总持币数错误: %v", currentAccount, err)
		return nil, err
	}
	totalFixReward, err := l.GetGroupTotalFixReward(currentAccount, latestSnapShotDay)
	if err != nil {
		log.Errorf("pos业务 - 根据地址 %s 获取团队总奖励数错误: %v", currentAccount, err)
		return nil, err
	}
	starLevel, _, _ := l.GetPosStarLevelFromConfig(currentAccount, currentHoldAmount, latestSnapShotDay)

	return &types.PosGroupInfo{
		Inviter:         l.MaskString(inviterNativeAccount),
		StarLevel:       starLevel,
		GroupHoldAmount: groupHoldAmount + currentHoldAmount,
		GroupFixReward:  totalFixReward,
	}, nil
}

// GetPosStarLevelFromConfig 根据地址得到StarLevel
func (l *PosSnapShotLogic) GetPosStarLevelFromConfig(owner string, amount float64, snapShotDay time.Time) (int32, float64, error) {
	config, exist := l.whiteList[owner]
	// 如果存在则以数据库配置的为主
	if exist {
		return config.StarLevel, config.Rate, nil
	} else {
		// 如果不存在则根据持币数量和下级数量来判断是否是星级用户
		var holdLevel, groupLevel int32 = 0, 0
		var holdLevelRate, groupLevelRate = 0.0, 0.0
		for _, rule := range l.starLevelRule {
			if amount >= rule.Amount {
				holdLevel = rule.StarLevel
				holdLevelRate = rule.Rate
				continue
			}
			break
		}
		// 如果持币量不够星级则直接返回星级为0
		if holdLevel == 0 {
			return 0, 0.0, nil
		}
		// 持币量达到星级后，再判断团队持币量
		groupHoldAmount, err := l.GetPosGroupHoldAmount(owner, snapShotDay)
		if err != nil {
			return 0, 0.0, nil
		}
		if holdLevel < 5 {
			// 如果持币量的星级小于5星，则判断团队持币星级的时候包含账户自己的持币量
			groupHoldAmount += amount
			for _, rule := range l.starLevelRule {
				if groupHoldAmount >= rule.GroupAmount {
					groupLevel = rule.StarLevel
					groupLevelRate = rule.Rate
					continue
				}
				break
			}
		} else {
			// 则先判断下级持币量是否达到了5星，如果达到5星则直接返回5星
			if groupHoldAmount >= l.starLevelRule[4].GroupAmount {
				groupLevel = l.starLevelRule[4].StarLevel
				groupLevelRate = l.starLevelRule[4].Rate
				return groupLevel, groupLevelRate, nil
			} else {
				// 如果下级持币量没有达到5星，则判断总持币量是否能达到最高4星
				groupHoldAmount += amount
				for _, rule := range l.starLevelRule {
					if groupHoldAmount >= rule.GroupAmount {
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
		return min(holdLevel, groupLevel), min(holdLevelRate, groupLevelRate), nil
	}
}

func (l *PosSnapShotLogic) GetInviterAccount(currentAccount string) (string, error) {
	var err error
	var record model.InviteRelation
	table := l.db.Table(model.TableNameInviteRelation)
	if err = table.Where("invitee = ?", currentAccount).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}

	return record.Inviter, nil
}

func (l *PosSnapShotLogic) MaskString(s string) string {
	// 检查字符串的长度
	if len(s) <= 10 {
		// 如果字符串长度小于等于10，则不需要掩码，直接返回原字符串
		return s
	}

	// 获取前5位和后5位
	start := s[:5]
	end := s[len(s)-5:]

	// 计算中间部分的长度
	middleLength := len(s) - 10                 // 总长度减去前5位和后5位
	middle := strings.Repeat("*", middleLength) // 用*替换中间部分

	// 返回组合后的字符串
	return start + middle + end
}
