package stake

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"math"
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/app/pos/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
	"time"
)

type StakeLogic struct {
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
	rewardConfig       *model.StakeRewardConfig
	leaderRewardConfig *model.StakeLeaderRewardConfig
	fixConfig          map[int32]model.StakeFixRateConfig
	totalLeaders       []model.StakeTotalLeader
}

func NewStakeLogic(ctx context.Context, srvCtx *svc.ServiceContext) *StakeLogic {
	return &StakeLogic{
		prefix:            "Stake质押Token业务 -",
		ctx:               ctx,
		srvCtx:            srvCtx,
		db:                srvCtx.DB,
		rd:                srvCtx.Redis,
		rs:                srvCtx.RedSync,
		rpcClient:         srvCtx.RpcClient,
		baseClient:        srvCtx.BaseClient,
		decimals:          uint8(srvCtx.TokenConfig.Decimals),
		LightHouseAddress: srvCtx.LightHouseAddress,

		// 业务相关参数
		rewardConfig:       srvCtx.StakeRewardConfig,
		leaderRewardConfig: srvCtx.LeaderRewardConfig,
		fixConfig:          srvCtx.StakeFixConfig,
		totalLeaders:       srvCtx.TotalAreaLeaders,
	}
}

// GetStakeMinAmount 获取最小质押金额
func (l *StakeLogic) GetStakeMinAmount(stakeType int64) interface{} {
	return l.fixConfig[int32(stakeType)]
}

// ProcessStakeToken 处理前端提交的质押交易
func (l *StakeLogic) ProcessStakeToken(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx) error {
	// 验证签名,Account0是交易的发起地址，同时也是转账token的地址，也是转sol到dex的地址，同时也是手续费的支付地址
	var prefix = "Stake业务 - 质押Token - "

	txIdStr := preCheckedTx.TxId.String()

	// 1. 发送给 base 模块签名 + 广播
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Pos", "StakeToken")
	if err != nil {
		log.Errorf("%s 发送交易失败: %v", prefix, err)
		return errors.New(utils.FilterAndTranslateSOLError(err))
	}
	if sentTxId == "" || sentTxId != txIdStr {
		log.Errorf("%s base返回的txId[%s]与预期[%s]不一致", prefix, sentTxId, txIdStr)
		return errors.New("sent transaction id not equal expected")
	}

	return nil
}

// ProcessUnStakeToken 处理前端提交的质押交易
func (l *StakeLogic) ProcessUnStakeToken(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx) error {
	// 验证签名,Account0是交易的发起地址，同时也是转账token的地址，也是转sol到dex的地址，同时也是手续费的支付地址
	var prefix = "Stake业务 - 解除质押Token - "

	txIdStr := preCheckedTx.TxId.String()

	// 1. 发送给 base 模块签名 + 广播
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Pos", "StakeToken")
	if err != nil {
		log.Errorf("%s 发送交易失败: %v", prefix, err)
		return errors.New(utils.FilterAndTranslateSOLError(err))
	}
	if sentTxId == "" || sentTxId != txIdStr {
		log.Errorf("%s base返回的txId[%s]与预期[%s]不一致", prefix, sentTxId, txIdStr)
		return errors.New("sent transaction id not equal expected")
	}

	return nil
}

// ProcessReStakeToken 处理前端提交的质押交易
func (l *StakeLogic) ProcessReStakeToken(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx) error {
	// 验证签名,Account0是交易的发起地址，同时也是转账token的地址，也是转sol到dex的地址，同时也是手续费的支付地址
	var prefix = "Stake业务 - 重新质押Token - "

	txIdStr := preCheckedTx.TxId.String()

	// 1. 发送给 base 模块签名 + 广播
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Pos", "StakeToken")
	if err != nil {
		log.Errorf("%s 发送交易失败: %v", prefix, err)
		return errors.New(utils.FilterAndTranslateSOLError(err))
	}
	if sentTxId == "" || sentTxId != txIdStr {
		log.Errorf("%s base返回的txId[%s]与预期[%s]不一致", prefix, sentTxId, txIdStr)
		return errors.New("sent transaction id not equal expected")
	}

	return nil
}

// HandleMarketBuyTx 处理交易所购买 token 的逻辑
func (l *StakeLogic) HandleMarketBuyTx(msg entity.NewScannedTx) error {
	txID := msg.TxSig.Signature.String()
	log.Infof("HandleMarketBuyTx: txId=%s", txID)

	insts := msg.DecodedTx.TransferCheckedInstructions
	if len(insts) == 0 {
		log.Warnf("HandleMarketBuyTx: no TransferChecked instructions in txId=%s, skipping", txID)
		return nil
	}

	toAccount := msg.DecodedTx.FromNativeAccount.String()
	slot := float64(msg.TxSig.Slot)

	var records []model.StakeBuyToken
	for _, inst := range insts {
		amount := float64(inst.Amount)
		records = append(records, model.StakeBuyToken{
			TxID:            txID,
			Slot:            slot,
			FromAccount:     inst.FromTokenAccount.String(),
			ToAccount:       toAccount,
			Amount:          amount,
			Locked:          false,
			StakedAmount:    0,
			RemainingAmount: amount,
		})
	}

	result := l.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tx_id"}},
		DoNothing: true,
	}).Create(&records)
	if result.Error != nil {
		return fmt.Errorf("HandleMarketBuyTx: insert t_stake_buy_token failed txId=%s: %w", txID, result.Error)
	}

	log.Infof("HandleMarketBuyTx: inserted %d records for txId=%s", len(records), txID)
	return nil
}

// HandleStakeTx 处理质押/解除质押/重新质押交易（SubService 均为 "StakeToken"）
func (l *StakeLogic) HandleStakeTx(msg entity.NewScannedTx) error {
	stakeTxSig := msg.TxSig.Signature
	stakeTxId := stakeTxSig.String()
	log.Infof("%s 收到 StakeToken 交易 txId=%s", l.prefix, stakeTxId)

	// 解析交易（支持 stake/unstake/restake 的不同指令数据长度）
	parser := NewStakeTxParser(l.rpcClient, float64(1000), "As9Z52f8Sioqr22KpS4xdzrhicwGwAu6x5SxVaHfvLws")
	parsedTx, err := parser.ParseStakeTx(context.Background(), stakeTxSig)
	if err != nil {
		log.Errorf("%s txId=%s 解析交易失败: %v", l.prefix, stakeTxId, err)
		return err
	}
	if !parsedTx.Success {
		log.Infof("%s txId=%s 链上执行失败，跳过处理", l.prefix, stakeTxId)
		return nil
	}

	disc := parsedTx.InstructionData.Discriminator
	staker := parsedTx.Accounts.Staker.String()

	switch {
	case bytes.Equal(disc, AnchorDiscriminator("stake")):
		log.Infof("%s txId=%s 识别为 stake 交易，staker=%s amount=%d stakeType=%d",
			l.prefix, stakeTxId, staker,
			parsedTx.InstructionData.Amount, parsedTx.InstructionData.StakeType)
		if err = l.handleStakeToken(parsedTx, parser.GetStakeAmount(parsedTx)); err != nil {
			log.Errorf("%s txId=%s 处理 stake 失败: %v", l.prefix, stakeTxId, err)
			return err
		}

	case bytes.Equal(disc, AnchorDiscriminator("unstake")):
		// stakeIndex 在 rawData[8]
		var stakeIndex uint8
		if len(parsedTx.RawData) > 8 {
			stakeIndex = parsedTx.RawData[8]
		}
		log.Infof("%s txId=%s 识别为 unstake 交易，staker=%s stakeIndex=%d rawData=%x",
			l.prefix, stakeTxId, staker, stakeIndex, parsedTx.RawData)

	case bytes.Equal(disc, AnchorDiscriminator("restake")):
		// stakeIndex 在 rawData[8]，stakeType 在 rawData[9]
		var stakeIndex, stakeType uint8
		if len(parsedTx.RawData) > 8 {
			stakeIndex = parsedTx.RawData[8]
		}
		if len(parsedTx.RawData) > 9 {
			stakeType = parsedTx.RawData[9]
		}
		log.Infof("%s txId=%s 识别为 restake 交易，staker=%s stakeIndex=%d stakeType=%d rawData=%x",
			l.prefix, stakeTxId, staker, stakeIndex, stakeType, parsedTx.RawData)

	default:
		log.Warnf("%s txId=%s 未知指令类型，discriminator=%x，跳过处理", l.prefix, stakeTxId, disc)
	}

	return nil
}

// handleStakeToken 处理质押成功后的逻辑
func (l *StakeLogic) handleStakeToken(parsedTx *types.ParsedStakeTx, stakeAmount uint64) error {
	dbTx := l.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
		}
	}()
	staker := parsedTx.Accounts.Staker.String()

	// 1. 将 ParsedStakeTx 写入质押记录，设置State为1（成功）
	_, err := l.createStakeRecord(dbTx, staker, stakeAmount, parsedTx)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("创建质押记录失败: %v", err)
	}

	// 2. 使用递归查询查找最近的区域领导
	directAreaLeaderAccount, err := l.findDirectLeader(dbTx, staker)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("查找邀请人失败: %v", err)
	}
	// 如果没有找到区域领导，直接提交事务结束
	if directAreaLeaderAccount == "" {
		log.Infof("%s 质押地址 %s 没有找到直接区域领导人", l.prefix, staker)
		return dbTx.Commit().Error
	}
	log.Infof("%s 质押地址 %s 直接区域领导人 %s", l.prefix, staker, directAreaLeaderAccount)

	// 3. 查询区域领导完整信息（Level、Leader等）
	areaLeader, err := l.getLeaderInfo(dbTx, directAreaLeaderAccount)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("查询区域领导信息失败: %v", err)
	}

	// 4. 查询质押者的总购买token剩余金额
	totalBuyTokenAmount, err := l.getBuyRemaining(dbTx, staker)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("查询总购买剩余金额失败: %v", err)
	}

	log.Infof("%s 质押地址 %s 可以用来抵扣的购买数量 %d 质押量 %d", l.prefix, staker, totalBuyTokenAmount, stakeAmount)

	// 如果总购买剩余金额为0，结束流程
	if totalBuyTokenAmount == 0 {
		return dbTx.Commit().Error
	}

	// 5. 计算基础奖励金额（取购买量与质押量的最小值）
	baseAmount := totalBuyTokenAmount
	if stakeAmount < totalBuyTokenAmount {
		baseAmount = stakeAmount
	}
	log.Infof("%s 地址[%s] 质押金额[%d] 基础奖励金额[%d] 区域领导[%s] level[%d]",
		l.prefix, staker, stakeAmount, baseAmount, areaLeader.NativeAccount, areaLeader.LeaderLevel)

	if baseAmount > 0 {
		// 扣减购买记录的RemainingAmount
		err = l.deductBuyRemaining(dbTx, staker, stakeAmount)
		if err != nil {
			dbTx.Rollback()
			return fmt.Errorf("%s 扣减购买记录剩余金额失败: %v", l.prefix, err)
		}
		err = l.rewardLeaders(dbTx, staker, areaLeader, baseAmount)
		if err != nil {
			dbTx.Rollback()
			return fmt.Errorf("%s 发放奖励给区域领导错误: %v", l.prefix, err)
		}
	}

	return dbTx.Commit().Error
}

// createStakeRecord 创建质押记录
func (l *StakeLogic) createStakeRecord(dbTx *gorm.DB, staker string, stakeAmount uint64, parsedTx *types.ParsedStakeTx) (*model.StakeRecord, error) {
	// 将 ParsedStakeTx 序列化为 JSON 存储
	parsedTxJSON, err := json.Marshal(parsedTx)
	if err != nil {
		return nil, fmt.Errorf("序列化交易数据失败: %v", err)
	}

	stakeRecord := &model.StakeRecord{
		Staker:      staker,
		StakeAmount: float64(stakeAmount),
		Status:      "success", // 直接设置为成功状态
		StakeTxHash: parsedTx.TransactionSignature.String(),
		UsedTxIds:   "[]",                 // 初始为空
		LockedTxIds: string(parsedTxJSON), // 存储完整的交易数据
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = dbTx.Table(model.TableNameStakeRecord).Create(stakeRecord).Error
	if err != nil {
		return nil, err
	}

	return stakeRecord, nil
}

// findDirectLeader 查找邀请过关系中离质押人最近的区域领导
func (l *StakeLogic) findDirectLeader(dbTx *gorm.DB, invitee string) (string, error) {
	// 先检查当前账户是否是区域领导
	var err error
	var areaLeader string
	// 使用更简单的查询方法
	err = dbTx.Table("public.t_stake_leader").
		Select("native_account").
		Where("native_account = ?", invitee).
		Scan(&areaLeader).Error

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Errorf("%s 查询直属区域领导错误: %v, SQL: SELECT NativeAccount FROM public.t_stake_area_leader WHERE NativeAccount = '%s'",
				l.prefix, err, invitee)
			return "", err
		}
	} else if areaLeader != "" {
		return areaLeader, nil
	}

	// 修正后的递归查询
	recursiveSQL := `
        with recursive invite_path as (
            -- 初始查询：检查直接邀请人是否是区域领导
            select 
                invitee,
                inviter,
                exists (
                    select 1 from public.t_stake_leader l 
                    where l.native_account = inviter
                ) as is_leader
            from public.t_invite_relation
            where invitee = ?
            
            union all
            
            -- 递归查询：如果上一级不是区域领导，继续向上查找
            select 
                ip.inviter,
                ir.inviter,
                exists (
                    select 1 from public.t_stake_leader l 
                    where l.native_account = ir.inviter
                ) as is_leader
            from public.t_invite_relation ir
            join invite_path ip on ir.invitee = ip.inviter
            where not ip.is_leader
        )
        select inviter 
        from invite_path 
        where is_leader = true 
        limit 1
    `

	err = dbTx.Raw(recursiveSQL, invitee).Row().Scan(&areaLeader)
	if err != nil {
		// 如果递归查询也没有找到区域领导，返回空字符串而不是错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}

	return areaLeader, nil
}

// getBuyRemaining 获取质押者的总购买token剩余金额
func (l *StakeLogic) getBuyRemaining(dbTx *gorm.DB, staker string) (uint64, error) {
	var totalAmount float64
	err := dbTx.Table(model.TableNameStakeBuyToken).
		Where("to_account= ?", staker).
		Select("coalesce(sum(remaining_amount), 0)").
		Scan(&totalAmount).Error

	if err != nil {
		return 0, err
	}

	return uint64(totalAmount), nil
}

// getLeaderInfo 查询区域领导完整信息
func (l *StakeLogic) getLeaderInfo(dbTx *gorm.DB, nativeAccount string) (*model.StakeLeader, error) {
	var leader model.StakeLeader
	err := dbTx.Table(model.TableNameStakeLeader).
		Where("native_account = ?", nativeAccount).
		First(&leader).Error
	if err != nil {
		return nil, err
	}
	return &leader, nil
}

// rewardLeaders 将区域经理奖励写入数据库，等待区域经理主动领取
// baseAmount = min(购买剩余量, 质押量)，各档奖励均基于 baseAmount 按比例计算
// level=2: 本人10% + 总区域领导共10%（按Share分配）
// level=1: 本人7% + 上级Leader3% + 总区域领导共10%（按Share分配）
func (l *StakeLogic) rewardLeaders(dbTx *gorm.DB, staker string, areaLeader *model.StakeLeader, baseAmount uint64) error {
	var records []model.StakeLeaderReward
	now := time.Now()

	switch areaLeader.LeaderLevel {
	case 2:
		// 无 Leader，直接区域经理获得 10%
		records = append(records, model.StakeLeaderReward{
			NativeAccount: areaLeader.NativeAccount,
			Staker:        staker,
			RewardType:    types.AreaLeaderRewardTypeDirect,
			BaseAmount:    float64(baseAmount),
			StakeShare:    10,
			RewardAmount:  float64(uint64(float64(baseAmount) * 0.10)),
			RewardState:   0,
			Pending:       false,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
	case 1:
		// 区域经理获得 7%
		records = append(records, model.StakeLeaderReward{
			NativeAccount: areaLeader.NativeAccount,
			Staker:        staker,
			RewardType:    types.AreaLeaderRewardTypeLevel1,
			BaseAmount:    float64(baseAmount),
			StakeShare:    7,
			RewardAmount:  float64(uint64(float64(baseAmount) * 0.07)),
			RewardState:   0,
			Pending:       false,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
		// 上级 Leader 获得 3%
		if areaLeader.UpLeader != "" {
			records = append(records, model.StakeLeaderReward{
				NativeAccount: areaLeader.UpLeader,
				Staker:        staker,
				RewardType:    types.AreaLeaderRewardTypeLeader,
				BaseAmount:    float64(baseAmount),
				StakeShare:    3,
				RewardAmount:  float64(uint64(float64(baseAmount) * 0.03)),
				RewardState:   0,
				Pending:       false,
				CreatedAt:     now,
				UpdatedAt:     now,
			})
		}
	default:
		log.Warnf("%s 区域领导 %s 的 level=%d 不在预期范围内，跳过本人/上级奖励", l.prefix, areaLeader.NativeAccount, areaLeader.LeaderLevel)
	}

	// 总区域领导（Share 存储为小数，如 0.07、0.03）
	for _, totalLeader := range l.totalLeaders {
		reward := uint64(float64(baseAmount) * totalLeader.StakeShare)
		if reward > 0 {
			records = append(records, model.StakeLeaderReward{
				NativeAccount: totalLeader.NativeAccount,
				Staker:        staker,
				RewardType:    types.AreaLeaderRewardTypeTotalLeader,
				BaseAmount:    float64(baseAmount),
				StakeShare:    totalLeader.StakeShare,
				RewardAmount:  float64(reward),
				RewardState:   0,
				Pending:       false,
				CreatedAt:     now,
				UpdatedAt:     now,
			})
		}
	}

	if len(records) == 0 {
		return nil
	}

	if err := dbTx.Table(model.TableNameStakeLeaderReward).Create(&records).Error; err != nil {
		return fmt.Errorf("写入区域经理奖励记录错误: %v", err)
	}
	log.Infof("%s 质押地址 %s 写入区域经理奖励记录 %d 条", l.prefix, staker, len(records))
	return nil
}

// deductBuyRemaining 扣减购买记录的剩余金额
func (l *StakeLogic) deductBuyRemaining(dbTx *gorm.DB, staker string, deductAmount uint64) error {
	// 查询质押者的购买记录，按Slot升序排列
	var buyRecords []model.StakeBuyToken
	err := dbTx.Table(model.TableNameStakeBuyToken).
		Where("to_account = ? and remaining_amount > 0", staker).
		Order("slot asc").
		Find(&buyRecords).Error
	if err != nil {
		return err
	}

	remainingDeduct := float64(deductAmount)

	// 按顺序扣减购买记录的剩余金额
	for _, record := range buyRecords {
		if remainingDeduct <= 0 {
			break
		}

		// 计算本次扣减的金额
		deductThisRecord := math.Min(record.RemainingAmount, remainingDeduct)

		// 更新购买记录的剩余金额
		newRemainingAmount := record.RemainingAmount - deductThisRecord
		err = dbTx.Table(model.TableNameStakeBuyToken).
			Where("tx_id = ?", record.TxID).
			Update("remaining_amount", newRemainingAmount).Error
		if err != nil {
			return err
		}

		remainingDeduct -= deductThisRecord
	}

	return nil
}

// GetUserAvailableStake 查询用户可用质押额度（排除锁定记录）
func (l *StakeLogic) GetUserAvailableStake(userAddress string) (uint64, error) {
	var totalAvailable struct {
		Total uint64 `gorm:"column:total_available"`
	}

	err := l.db.Table(model.TableNameStakeBuyToken).
		Select("coalesce(sum(remaining_amount), 0) as total_available").
		Where("to_account= ? and locked = ?", userAddress, false).
		Scan(&totalAvailable).Error

	if err != nil {
		return 0, fmt.Errorf("查询可用质押额度失败: %v", err)
	}

	return totalAvailable.Total, nil
}

// GetLeaderInfo 根据 native account 查询区域经理信息，不存在时返回 nil
func (l *StakeLogic) GetLeaderInfo(nativeAccount string) (*model.StakeLeader, error) {
	var leader model.StakeLeader
	err := l.db.Table(model.TableNameStakeLeader).
		Where("native_account = ?", nativeAccount).
		First(&leader).Error
	if err != nil {
		return nil, err
	}
	return &leader, nil
}

func (l *StakeLogic) GetLeaderRewards(nativeAccount string) ([]model.StakeLeaderReward, error) {
	var rewards []model.StakeLeaderReward
	err := l.db.Table(model.TableNameStakeLeaderReward).
		Where("native_account = ? and state = ? and pending = ?", nativeAccount, 0, false).
		Order("created_at asc").
		Find(&rewards).Error
	if err != nil {
		return nil, err
	}
	return rewards, nil
}

func (l *StakeLogic) GetLeaderTxInfo(account solana.PublicKey) (*types.LeaderRewardTxInfo, error) {
	// 查询未领取的奖励总额
	var totalReward float64
	err := l.db.Table(model.TableNameStakeLeaderReward).
		Select("coalesce(sum(reward_amount), 0)").
		Where("native_account = ? and reward_state = ? and pending = ?", account.String(), 0, false).
		Scan(&totalReward).Error
	if err != nil {
		return nil, fmt.Errorf("查询区域经理未领取奖励总额错误: %v", err)
	}

	return &types.LeaderRewardTxInfo{
		RewardAccount: l.leaderRewardConfig.RewardAccount,
		Mint:          l.srvCtx.TokenConfig.Mint,
		Decimals:      l.srvCtx.TokenConfig.Decimals,
		TotalReward:   totalReward,
	}, nil
}
