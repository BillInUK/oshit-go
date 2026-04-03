package lottery

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
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
	"time"
)

type LotteryLogic struct {
	prefix          string
	ctx             context.Context
	srvCtx          *svc.ServiceContext
	db              *gorm.DB
	rd              redis.UniversalClient
	rs              redsync.Redsync
	rpcClient       *rpc.Client
	baseClient      *rewardrpc.BaseClient
	takeTokenConfig *model.TakeTokenConfig
}

func NewLotteryLogic(ctx context.Context, srvCtx *svc.ServiceContext) *LotteryLogic {
	return &LotteryLogic{
		prefix:          "Lottery",
		ctx:             ctx,
		srvCtx:          srvCtx,
		db:              srvCtx.DB,
		rd:              srvCtx.Redis,
		rs:              srvCtx.RedSync,
		rpcClient:       srvCtx.RpcClient,
		baseClient:      srvCtx.BaseClient,
		takeTokenConfig: srvCtx.TakeTokenConfig,
	}
}

// isLotteryThreshold 判断 take_count 是否触发抽奖阈值
func isLotteryThreshold(takeCount int32) bool {
	return takeCount == 5 || takeCount == 10 || takeCount == 20
}

// GetStatus 查询今日的领取统计
func (l *LotteryLogic) GetStatus(nativeAccount string) (*model.DailyClaimStats, error) {
	today := time.Now().Truncate(24 * time.Hour)
	var stats model.DailyClaimStats
	err := l.db.Table(model.TableNameDailyClaimStats).
		Where("native_account = ? AND take_date = ?", nativeAccount, today).
		First(&stats).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// ExecuteLottery 执行抽奖：校验条件，创建抽奖记录，递增 lottery_count
func (l *LotteryLogic) ExecuteLottery(nativeAccount string) (*model.LotteryReward, error) {
	prefix := fmt.Sprintf("%s ExecuteLottery account=%s -", l.prefix, nativeAccount)

	today := time.Now().Truncate(24 * time.Hour)

	// 1. 查询今日统计
	var stats model.DailyClaimStats
	err := l.db.Table(model.TableNameDailyClaimStats).
		Where("native_account = ? AND take_date = ?", nativeAccount, today).
		First(&stats).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("no daily claim stats found, cannot execute lottery")
	}
	if err != nil {
		log.Errorf("%s 查询每日统计错误: %v", prefix, err)
		return nil, errors.New("query daily claim stats error")
	}

	// 2. 校验条件
	if !stats.NeedLottery {
		return nil, errors.New("lottery is not required")
	}
	if !isLotteryThreshold(stats.TakeCount) {
		return nil, errors.New("take count does not meet lottery threshold")
	}
	if stats.LotteryCount >= 3 {
		return nil, errors.New("lottery count has reached the maximum limit of 3")
	}

	// 3. 检查是否存在是否存在未兑换的抽奖记录
	var pendingReward model.LotteryReward
	pendingErr := l.db.Table(model.TableNameLotteryReward).
		Where("native_account = ? AND pending = ? AND state = ?", nativeAccount, false, 0).
		First(&pendingReward).Error
	if pendingErr == nil {
		// 已有pending记录
		return nil, errors.New("there is already a pending lottery reward, please claim it first")
	}
	if !errors.Is(pendingErr, gorm.ErrRecordNotFound) {
		log.Errorf("%s 查询pending抽奖记录错误: %v", prefix, pendingErr)
		return nil, errors.New("query pending lottery reward error")
	}

	// 4. 生成随机奖励金额（字面值转 raw amount）
	rewardAmountLiteral, err := GenerateWeightedLotteryAmount(int(stats.TakeCount))
	if err != nil {
		log.Errorf("%s 生成随机奖励金额错误: %v", prefix, err)
		return nil, errors.New("generate lottery amount error")
	}
	rewardAmountRaw := float64(rewardAmountLiteral) * l.srvCtx.TokenDecimal

	// 5. 创建抽奖记录
	reward := model.LotteryReward{
		NativeAccount: nativeAccount,
		RewardAmount:  rewardAmountRaw,
		RewardType:    0,
		State:         0,
		Pending:       true,
		RewardDay:     today,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	dbTx := l.db.Begin()
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}

	if err := dbTx.Table(model.TableNameLotteryReward).Omit("record_id").Create(&reward).Error; err != nil {
		dbTx.Rollback()
		log.Errorf("%s 创建抽奖记录错误: %v", prefix, err)
		return nil, errors.New("create lottery reward record error")
	}

	// 6. 递增 lottery_count
	if err := dbTx.Table(model.TableNameDailyClaimStats).
		Where("native_account = ? AND take_date = ?", nativeAccount, today).
		UpdateColumn("lottery_count", gorm.Expr("lottery_count + 1")).Error; err != nil {
		dbTx.Rollback()
		log.Errorf("%s 递增 lottery_count 错误: %v", prefix, err)
		return nil, errors.New("update lottery_count error")
	}

	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s 提交事务错误: %v", prefix, err)
		return nil, errors.New("commit transaction error")
	}

	return &reward, nil
}

// GetUnclaimedRewards 查询未领取的抽奖奖励 (pending=true, state=0)
func (l *LotteryLogic) GetUnclaimedRewards(nativeAccount string) ([]*model.LotteryReward, error) {
	var rewards []*model.LotteryReward
	err := l.db.Table(model.TableNameLotteryReward).
		Where("native_account = ? AND pending = ? AND state = ?", nativeAccount, false, 0).
		Find(&rewards).Error
	if err != nil {
		return nil, err
	}
	return rewards, nil
}

// GetTxInfo 获取抽奖领取的交易信息
func (l *LotteryLogic) GetTxInfo(ctx context.Context, recordId string) (*types.ClaimLotteryTxInfo, error) {
	prefix := fmt.Sprintf("%s GetTxInfo recordId=%s -", l.prefix, recordId)

	// 通过 recordId 查找 pending=true, state=0 的奖励记录
	var reward model.LotteryReward
	err := l.db.Table(model.TableNameLotteryReward).
		Where("record_id = ? AND pending = ? AND state = ?", recordId, true, 0).
		First(&reward).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("no pending lottery reward found")
	}
	if err != nil {
		log.Errorf("%s 查询pending奖励记录错误: %v", prefix, err)
		return nil, errors.New("query lottery reward error")
	}

	// 计算 CostFee：基于 take token config 的默认奖励金额
	quoteSOLPrice, err := l.baseClient.GetTokenQuoteSOLPrice(ctx)
	if err != nil {
		log.Errorf("%s 获取 token/SOL 价格失败: %v", prefix, err)
		return nil, errors.New("get token quote sol price failed")
	}
	configAmount := float64(l.takeTokenConfig.Amount)
	costFee := quoteSOLPrice * configAmount / l.srvCtx.TokenDecimal * float64(solana.LAMPORTS_PER_SOL)

	txInfo := &types.ClaimLotteryTxInfo{
		RecordId:      reward.RecordID,
		RewardAccount: l.srvCtx.TakeTokenConfig.RewardAccount,
		Mint:          l.srvCtx.TokenConfig.Mint,
		CostAccount:   l.takeTokenConfig.CostAccount,
		Decimals:      int32(l.srvCtx.TokenConfig.Decimals),
		LotteryAmount: reward.RewardAmount,
		CostFee:       costFee,
	}

	return txInfo, nil
}

// recordLotteryClaim 创建 t_lottery_claim_record 和 t_service_tx（在同一事务中）
func (l *LotteryLogic) recordLotteryClaim(dbTx *gorm.DB, nativeAccount, txId, rewardId string) error {
	// 创建 service_tx 记录
	txRecord := model.ServiceTx{
		Service:    "Reward",
		SubService: "Lottery",
		TxID:       txId,
		CreatedAt:  time.Now(),
	}
	if err := dbTx.Table(model.TableNameServiceTx).Create(&txRecord).Error; err != nil {
		return fmt.Errorf("insert service tx error: %v", err)
	}

	// 创建 lottery_claim_record 记录
	claimRecord := model.LotteryClaimRecord{
		RewardIds: fmt.Sprintf("{%s}", rewardId),
		TxID:      txId,
		State:     0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := dbTx.Table(model.TableNameLotteryClaimRecord).Omit("record_id").Create(&claimRecord).Error; err != nil {
		return fmt.Errorf("insert lottery claim record error: %v", err)
	}

	return nil
}

// ProcessCommitTx 处理前端提交的抽奖领取交易
func (l *LotteryLogic) ProcessCommitTx(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx, rewardId string) (*solana.Signature, error) {
	prefix := fmt.Sprintf("%s ProcessCommitTx from=%s rewardId=%s -", l.prefix, preCheckedTx.From.String(), rewardId)

	// 查询抽奖记录
	var reward model.LotteryReward
	if err := l.db.Table(model.TableNameLotteryReward).
		Where("record_id = ? AND native_account = ? AND pending = ? AND state = ?", rewardId, preCheckedTx.From.String(), true, 0).
		First(&reward).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("lottery reward record not found or already claimed")
		}
		log.Errorf("%s 查询抽奖记录错误: %v", prefix, err)
		return nil, errors.New("query lottery reward error")
	}

	// 获取交易信息用于校验
	rewardInfo, err := l.GetTxInfo(ctx, reward.RecordID)
	if err != nil {
		log.Errorf("%s 获取交易信息错误: %v", prefix, err)
		return nil, fmt.Errorf("get lottery tx info error: %v", err)
	}

	// 解析交易
	decodedTx, err := l.decodeSOLTx(&preCheckedTx.SOLTx, rewardInfo)
	if err != nil {
		log.Errorf("%s 解析solana交易错误: %v", prefix, err)
		return nil, fmt.Errorf("decode transaction error: %v", err)
	}

	// 校验交易
	if err := l.checkSOLTx(decodedTx, rewardInfo); err != nil {
		log.Errorf("%s 校验交易错误: %v", prefix, err)
		return nil, fmt.Errorf("check transaction error: %v", err)
	}

	// 发送交易给 base 模块
	txIdStr, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Reward", "Lottery")
	if err != nil {
		log.Errorf("%s 发送交易失败,错误: %v", prefix, err)
		return nil, errors.New(utils.FilterAndTranslateSOLError(err))
	}
	txId := solana.MustSignatureFromBase58(txIdStr)

	// 在事务中记录领取记录
	dbTx := l.db.Begin()
	if dbTx.Error != nil {
		return &txId, errors.New("begin transaction error")
	}

	if err := l.recordLotteryClaim(dbTx, preCheckedTx.From.String(), txIdStr, rewardId); err != nil {
		dbTx.Rollback()
		log.Errorf("%s 记录领取记录错误: %v", prefix, err)
		return &txId, errors.New("record lottery claim error")
	}

	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s 提交事务错误: %v", prefix, err)
		return &txId, errors.New("commit transaction error")
	}

	return &txId, nil
}
