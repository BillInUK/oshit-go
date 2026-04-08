package stake

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/app/pos/api/types"
	"oshit-go/common/pkg/dal/model"
	"time"
)

type StakeRewardLogic struct {
	prefix   string
	decimals uint8
	ctx      context.Context
	srvCtx   *svc.ServiceContext
	db       *gorm.DB
	rd       redis.UniversalClient
	rs       redsync.Redsync

	rpcClient         *rpc.Client
	LightHouseAddress solana.PublicKey
	baseClient        *posrpc.BaseClient

	// 业务相关参数
	distLevel     int32                               // 向上递归奖励的层级
	serviceConfig *model.StakeRewardConfig            // 下发奖励配置
	starWhiteList map[string]model.StakeStarWhitelist // 星级用户白名单
	starLevelRule map[int32]model.StakeStarLevelRule  // 星级评定规则
}

func NewStakeRewardLogic(ctx context.Context, srvCtx *svc.ServiceContext) *StakeRewardLogic {
	return &StakeRewardLogic{
		prefix:            "TakeToken业务 -",
		ctx:               ctx,
		srvCtx:            srvCtx,
		db:                srvCtx.DB,
		rd:                srvCtx.Redis,
		rs:                srvCtx.RedSync,
		rpcClient:         srvCtx.RpcClient,
		baseClient:        srvCtx.BaseClient,
		distLevel:         srvCtx.StakeDistLevel,
		starWhiteList:     srvCtx.StakeStarWhitelist,
		starLevelRule:     srvCtx.StakeStarLevelRule,
		decimals:          uint8(srvCtx.TokenConfig.Decimals),
		LightHouseAddress: srvCtx.LightHouseAddress,
	}
}

// GetConfig 获取奖励规则配置
func (l *StakeRewardLogic) GetConfig() (*model.StakeRewardConfig, error) {
	return l.serviceConfig, nil
}

// GetClaimRecord 根据交易id获取领取stake奖励记录
func (l *StakeRewardLogic) GetClaimRecord(txId string) (*model.StakeRewardClaimRecord, error) {
	var err error
	var record model.StakeRewardClaimRecord
	table := l.db.Table(model.TableNameStakeRewardClaimRecord)
	if err = table.Where("tx_id = ?", txId).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetRewardRecord 根据地址获取未领取的stake奖励
func (l *StakeRewardLogic) GetRewardRecord(nativeAccount string) ([]model.StakeReward, error) {
	var rewards []model.StakeReward

	table := l.db.Table(model.TableNameStakeReward)
	query := table.Where("native_account = ? and state = ? and pending  = ?", nativeAccount, 0, false)

	// 查询数据并获取总数
	if err := query.Order("created_at desc").Find(&rewards).Error; err != nil {
		return nil, err
	}
	// 返回结果和总数
	return rewards, nil
}

func (l *StakeRewardLogic) GetStakeSnapShot(nativeAccount string, snapShotDay time.Time) ([]model.StakeSnapShot, error) {
	var snapShots []model.StakeSnapShot
	if err := l.db.Table(model.TableNameStakeSnapShot).
		Where("native_account = ? and snap_day = ?", nativeAccount, snapShotDay).
		Scan(&snapShots).Error; err != nil {
		log.Errorf("%s 根据日期查询快照失败: %v", l.prefix, err)
		return nil, err
	}
	return snapShots, nil
}

// GetStakeStarLevelFromConfig 评定质押星级
func (l *StakeRewardLogic) GetStakeStarLevelFromConfig(owner string, amount float64, snapShotDay time.Time) (int32, float64, error) {
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

func (l *StakeRewardLogic) getGroupStakeAmount(rootAccount string, snapShotDay time.Time) (float64, error) {
	var totalHoldAmount float64
	query := `
		WITH RECURSIVE invite_tree AS (
			SELECT
				inviter AS account
			FROM
				public.t_invite_relation
			WHERE
				inviter = ?  -- 输入邀请人的账户
			UNION
			SELECT
				r.invitee
			FROM
				public.t_invite_relation r
			INNER JOIN
				invite_tree it
			ON
				r.inviter = it.account
		),
		group_accounts AS (
			SELECT DISTINCT(account)
			FROM invite_tree
			WHERE account != ?  -- 排除掉邀请人自己的账户
		),
		filtered_snap_shot AS (
			SELECT *
			FROM t_stake_snap_shot
			WHERE snap_day = CAST(? AS DATE)
		)
		SELECT
			COALESCE(SUM(fss.amount), 0) AS totalHoldAmount
		FROM
			group_accounts gp
		LEFT JOIN
			filtered_snap_shot fss
		ON
			gp.account = fss.native_account;
	`
	err := l.db.Raw(query, rootAccount, rootAccount, snapShotDay).Scan(&totalHoldAmount).Error
	if err != nil {
		return 0, err
	}

	return totalHoldAmount, nil
}

// GetStakeRewardStat 获取stake奖励明细
func (l *StakeRewardLogic) GetStakeRewardStat(nativeAccount string) (*types.StakeRewardStat, error) {
	var err error
	var stat types.StakeRewardStat
	snapShotDay := time.Now()

	// 查询当日质押量
	stakeSnapShots, err := l.GetStakeSnapShot(nativeAccount, snapShotDay)
	if err != nil {
		log.Errorf("%s 查询地址 %s 快照错误: %v", l.prefix, nativeAccount, err)
		return nil, err
	}
	stakeAmount := 0.0
	for _, s := range stakeSnapShots {
		stakeAmount += s.Amount
	}

	// 直接使用 DB 实例，每次查询都重新指定表名
	starLevel, _, _ := l.GetStakeStarLevelFromConfig(nativeAccount, stakeAmount, time.Now())

	// 查询每日固定利息累计收入
	accumulatedInterest := 0.0
	err = l.db.Table(model.TableNameStakeReward).
		Select(`COALESCE(SUM(reward_amount), 0) AS reward_amount`).
		Where(`native_account = ? and reward_type = ? and state = 0 and pending=false`, nativeAccount, types.StakeFixed).
		Scan(&accumulatedInterest).Error
	if err != nil {
		return nil, err
	}

	// 查询质押邀请奖励
	inviteReward := 0.0
	err = l.db.Table(model.TableNameStakeReward).
		Select(`reward_amount`).
		Where(`native_account = ? and reward_type = ? and state = 0 and pending=false`, nativeAccount, types.StakeInvite).
		Order("day desc").
		Limit(1).
		Scan(&inviteReward).Error
	if err != nil {
		return nil, err
	}

	// 查询质押激励奖励个人星级部分
	starReward := 0.0
	err = l.db.Table(model.TableNameStakeReward).
		Select(`reward_amount`).
		Where(`native_account = ? and reward_type = ? and state = 0 and pending=false`, nativeAccount, types.StakeStar).
		Order("day desc").
		Limit(1).
		Scan(&starReward).Error
	if err != nil {
		return nil, err
	}

	// 查询团队极差奖励
	groupReward := 0.0
	err = l.db.Table(model.TableNameStakeReward).
		Select(`reward_amount`).
		Where(`native_account = ? and reward_type = ? and state = 0 and pending=false`, nativeAccount, types.StakeStarGroup).
		Order("day desc").
		Limit(1).
		Scan(&groupReward).Error
	if err != nil {
		return nil, err
	}

	// 查询团队总质押量
	groupTotalStake, _ := l.getGroupStakeAmount(nativeAccount, snapShotDay)
	if err != nil {
		log.Errorf("%s 查询地址 %s 团队质押总奖励错误: %v", l.prefix, nativeAccount, err)
		return nil, err
	}

	// 查询总奖励 = 质押邀请奖励 + 激励奖励个人星级部分 + 质押激励奖励团队部分
	totalReward := inviteReward + starReward + groupReward

	log.Infof("查询地址: %s 日期 %v 奖励: StarLevel: %d StakeAmount: %.3f AccumulatedInterest: %.3f InviteReward: %.3f StarReward %.3f GroupReward %.3f TotalReward %.3f GroupTotalStake %.3f",
		nativeAccount, snapShotDay, starLevel, stakeAmount, accumulatedInterest, inviteReward, starReward, groupReward, totalReward, groupTotalStake)

	stat.StarLevel = starLevel
	stat.StakeAmount = stakeAmount
	stat.AccumulatedInterest = accumulatedInterest
	stat.InviteReward = inviteReward
	stat.StarReward = starReward
	stat.GroupReward = groupReward
	stat.TotalReward = totalReward
	stat.GroupTotalStake = groupTotalStake

	return &stat, nil
}

// GetTxInfo 根据交易id获取领取stake奖励记录
func (l *StakeRewardLogic) GetTxInfo(nativeAccount string) (*types.ClaimStakeRewardTxInfo, error) {
	return nil, nil
}
