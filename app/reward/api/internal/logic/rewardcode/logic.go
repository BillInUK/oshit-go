package rewardcode

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	rewardrpc "oshit-go/app/reward/api/internal/rpc"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/app/reward/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
	"time"
)

type RewardCodeLogic struct {
	prefix           string
	ctx              context.Context
	srvCtx           *svc.ServiceContext
	db               *gorm.DB
	rd               redis.UniversalClient
	rs               redsync.Redsync
	rpcClient        *rpc.Client
	baseClient       *rewardrpc.BaseClient
	rewardCodeConfig *model.RewardCodeConfig
	service          constants.ServiceName
	subService       constants.SubServiceName
}

func NewRewardCodeLogic(ctx context.Context, srvCtx *svc.ServiceContext) *RewardCodeLogic {
	return &RewardCodeLogic{
		prefix:           "RewardCode业务 -",
		ctx:              ctx,
		srvCtx:           srvCtx,
		db:               srvCtx.DB,
		rd:               srvCtx.Redis,
		rs:               srvCtx.RedSync,
		rpcClient:        srvCtx.RpcClient,
		baseClient:       srvCtx.BaseClient,
		rewardCodeConfig: srvCtx.RewardCodeConfig,
		service:          constants.ServiceReward,
		subService:       constants.SubServiceRewardCode,
	}
}

// GetInfo 根据奖励码查询奖励码基本信息
func (l *RewardCodeLogic) GetInfo(rewardCode string) (*model.RewardCode, error) {
	var rc model.RewardCode
	err := l.db.Table(model.TableNameRewardCode).
		Where("reward_code = ?", rewardCode).
		First(&rc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("reward code not found")
	}
	if err != nil {
		return nil, errors.New("query reward code error")
	}
	return &rc, nil
}

// GetTxInfo 根据奖励码获取交易参数信息
func (l *RewardCodeLogic) GetTxInfo(ctx context.Context, rewardCode string) (*types.RewardCodeTxInfo, error) {
	prefix := fmt.Sprintf("%s GetTxInfo rewardCode=%s -", l.prefix, rewardCode)

	// 1. 根据奖励码查询 t_reward_code，要求 tx_state=0（未使用）
	var rc model.RewardCode
	err := l.db.Table(model.TableNameRewardCode).
		Where("reward_code = ?", rewardCode).
		First(&rc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("reward code not found or already used")
	}
	if err != nil {
		log.Errorf("%s 查询奖励码错误: %v", prefix, err)
		return nil, errors.New("query reward code error")
	}

	// 2. 检查奖励码是否已过期
	if rc.RewardState < 0 || rc.ExpiredAt.Before(time.Now()) {
		return nil, errors.New("reward code has expired")
	}

	// 3. 根据 reward_amount 查询 t_reward_code_fee 获取 cost_rate
	var feeConfig model.RewardCodeFee
	err = l.db.Table(model.TableNameRewardCodeFee).
		Where("amount = ?", rc.RewardAmount).
		First(&feeConfig).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("reward code fee config not found for this amount")
	}
	if err != nil {
		log.Errorf("%s 查询奖励码费率配置错误: %v", prefix, err)
		return nil, errors.New("query reward code fee config error")
	}

	// 4. 从 base 模块获取 token/SOL 兑换价格
	quoteSOLPrice, err := l.baseClient.GetTokenQuoteSOLPrice(ctx)
	if err != nil {
		log.Errorf("%s 获取 token/SOL 价格失败: %v", prefix, err)
		return nil, errors.New("get token quote sol price failed")
	}

	// 5. 计算 cost_fee (lamports): reward_amount / tokenDecimal * quoteSOLPrice * costRate * LAMPORTS_PER_SOL
	costFee := rc.RewardAmount / l.srvCtx.TokenDecimal * quoteSOLPrice * feeConfig.FeeRate / 100 * float64(solana.LAMPORTS_PER_SOL)

	txInfo := &types.RewardCodeTxInfo{
		RewardAccount: l.rewardCodeConfig.RewardAccount,
		Mint:          l.srvCtx.TokenConfig.Mint,
		Decimals:      int32(l.srvCtx.TokenConfig.Decimals),
		CostAccount:   l.rewardCodeConfig.CostAccount,
		RewardAmount:  rc.RewardAmount,
		CostFee:       costFee,
	}

	return txInfo, nil
}

// ProcessCommitTx 处理前端提交的奖励码领取交易
func (l *RewardCodeLogic) ProcessCommitTx(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx, rewardCode string) error {
	txIdStr := preCheckedTx.TxId.String()
	prefix := fmt.Sprintf("%s 处理用户提交交易Id %v rewardCode=%s -", l.prefix, txIdStr, rewardCode)

	// 1. 分布式锁，防止同一奖励码被重复处理
	mutex := l.rs.NewMutex("reward-code:process:commit-tx:"+rewardCode, redsync.WithExpiry(1*time.Hour))
	if err := mutex.Lock(); err != nil {
		var errTaken *redsync.ErrTaken
		if !errors.As(err, &errTaken) {
			log.Errorf("%s 获取处理交易的分布式锁错误: %v", prefix, err)
			return errors.New("process transaction error")
		}
		return errors.New("reward code is being processed, please try again later")
	}
	defer mutex.Unlock()

	// 2. 再次查询奖励码确认仍可用（锁内二次校验）
	var rc model.RewardCode
	err := l.db.Table(model.TableNameRewardCode).
		Where("reward_code = ?", rewardCode).
		First(&rc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("reward code not found or already used")
	}
	if err != nil {
		log.Errorf("%s 查询奖励码错误: %v", prefix, err)
		return errors.New("query reward code error")
	}
	if rc.RewardState < 0 || rc.ExpiredAt.Before(time.Now()) {
		return errors.New("reward code has expired")
	}

	// 3. 获取交易信息用于校验
	txInfo, err := l.GetTxInfo(ctx, rewardCode)
	if err != nil {
		log.Errorf("%s 获取交易信息错误: %v", prefix, err)
		return fmt.Errorf("get reward code tx info error: %v", err)
	}

	// 4. 解析交易
	decodedTx, err := l.decodeSOLTx(&preCheckedTx.SOLTx, txInfo)
	if err != nil {
		log.Errorf("%s 解析solana交易错误: %v", prefix, err)
		return fmt.Errorf("decode transaction error: %v", err)
	}

	// 5. 校验交易
	if err := l.checkSOLTx(decodedTx, txInfo); err != nil {
		log.Errorf("%s 校验交易错误: %v", prefix, err)
		return fmt.Errorf("check transaction error: %v", err)
	}

	// 6. 发送交易给 base 模块
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, l.service, l.subService)
	if err != nil {
		log.Errorf("%s 发送交易失败,错误: %v", prefix, err)
		return errors.New(utils.FilterAndTranslateSOLError(err))
	}
	if sentTxId == "" || sentTxId != txIdStr {
		log.Errorf("%s 发送的交易Id %s 与预期的不一致", prefix, sentTxId)
		return errors.New("sent transaction id not equal expected")
	}

	// 7. 记录交易 ID 到 t_reward_code（tx_state 保持 0，等待 kafka 确认后更新）
	if err := l.db.Table(model.TableNameRewardCode).
		Where("record_id = ?", rc.RecordID).
		Updates(map[string]interface{}{"native_account": decodedTx.FromNativeAccount, "tx_id": txIdStr}).Error; err != nil {
		log.Errorf("%s 更新奖励码 tx_id 错误: %v", prefix, err)
		return errors.New("record reward code tx id error")
	}

	return nil
}
