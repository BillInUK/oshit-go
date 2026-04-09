package stake

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
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/app/pos/api/internal/svc"
	"oshit-go/app/pos/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
	"strings"
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

	// 快照逻辑
	snapShotLogic *StakeSnapShotLogic
}

func NewStakeRewardLogic(ctx context.Context, srvCtx *svc.ServiceContext) *StakeRewardLogic {
	return &StakeRewardLogic{
		prefix:            "Stake奖励业务 -",
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
		snapShotLogic:     NewStakeSnapShotLogic(ctx, srvCtx),
	}
}

// GetConfig 获取奖励规则配置
func (l *StakeRewardLogic) GetConfig() (*model.StakeRewardConfig, error) {
	return l.serviceConfig, nil
}

// GetClaimRecord 根据交易id获取领取stake奖励记录
func (l *StakeRewardLogic) GetClaimRecord(txId string) (*model.StakeRewardClaim, error) {
	var err error
	var record model.StakeRewardClaim
	table := l.db.Table(model.TableNameStakeRewardClaim)
	if err = table.Where("tx_id = ?", txId).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetRewardRecord 根据地址获取未领取的stake奖励
func (l *StakeRewardLogic) GetRewardRecord(nativeAccount string) ([]model.StakeReward, error) {
	var rewards []model.StakeReward

	table := l.db.Table(model.TableNameStakeReward)
	query := table.Where("native_account = ? and reward_state = ? and pending  = ?", nativeAccount, 0, false)

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

// GetTxInfo 获取领取 stake 奖励的交易参数
func (l *StakeRewardLogic) GetTxInfo(nativeAccount string) (*types.ClaimStakeRewardTxInfo, error) {
	config := l.srvCtx.StakeRewardConfig
	if config == nil {
		return nil, fmt.Errorf("stake reward config not loaded")
	}

	// 查询所有可领取奖励（reward_state=0, pending=false）
	var rewards []model.StakeReward
	if err := l.db.Table(model.TableNameStakeReward).
		Where("native_account = ? AND reward_state = ? AND pending = ?", nativeAccount, 0, false).
		Find(&rewards).Error; err != nil {
		return nil, fmt.Errorf("query stake rewards error: %v", err)
	}
	if len(rewards) == 0 {
		return nil, fmt.Errorf("no claimable stake rewards found")
	}

	// 累加总奖励金额（raw amount，含精度）
	var totalReward float64
	for _, r := range rewards {
		totalReward += r.RewardAmount
	}

	// 获取 token/SOL 价格
	quoteSOLPrice, err := l.baseClient.GetTokenQuoteSOLPrice(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("get token quote sol price error: %v", err)
	}

	// 成本费用：基于配置固定金额 QuoteTokenAmount 计算（不随本次奖励金额变化）
	costFee := quoteSOLPrice * config.QuoteTokenAmount / l.srvCtx.TokenDecimal * float64(solana.LAMPORTS_PER_SOL)
	// 用户本次奖励对应的 SOL 报价（仅供前端展示参考）
	quoteAmount := quoteSOLPrice * totalReward / l.srvCtx.TokenDecimal * float64(solana.LAMPORTS_PER_SOL)

	return &types.ClaimStakeRewardTxInfo{
		RewardAccount: config.RewardAccount,
		Mint:          l.srvCtx.TokenConfig.Mint,
		Decimals:      int32(l.srvCtx.TokenConfig.Decimals),
		TotalReward:   totalReward,
		QuoteAmount:   quoteAmount,
		CostAccount:   config.CostAccount,
		CostFeeRate:   config.CostFeeRate,
		CostFee:       costFee,
	}, nil
}

// recordStakeClaim 在同一数据库事务中：
// 1. 将本次可领取奖励标记为 pending=true
// 2. 写入 t_service_tx
// 3. 写入 t_stake_reward_claim_record
func (l *StakeRewardLogic) recordStakeClaim(nativeAccount, txId string, rewards []model.StakeReward) error {
	dbTx := l.db.Begin()
	if dbTx.Error != nil {
		return fmt.Errorf("begin transaction error: %v", dbTx.Error)
	}
	defer dbTx.Rollback()

	// 收集所有 rewardId，格式化为 PostgreSQL array 字符串 {id1,id2,...}
	rewardIds := make([]string, 0, len(rewards))
	for _, r := range rewards {
		rewardIds = append(rewardIds, r.RecordID)
	}
	rewardIdsStr := "{" + strings.Join(rewardIds, ",") + "}"

	// 1. 标记奖励为 pending
	if err := dbTx.Table(model.TableNameStakeReward).
		Where("native_account = ? AND reward_state = ? AND pending = ?", nativeAccount, 0, false).
		Updates(map[string]interface{}{
			"pending":    true,
			"tx_id":      txId,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("mark stake rewards as pending error: %v", err)
	}

	// 2. 写入 t_stake_reward_claim_record
	claimRecord := model.StakeRewardClaim{
		RewardIds: rewardIdsStr,
		TxID:      txId,
		TxState:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := dbTx.Table(model.TableNameStakeRewardClaim).Omit("record_id").Create(&claimRecord).Error; err != nil {
		return fmt.Errorf("insert stake reward claim record error: %v", err)
	}

	if err := dbTx.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction error: %v", err)
	}
	return nil
}

// ProcessCommitTx 处理前端提交的 stake 奖励领取交易
func (l *StakeRewardLogic) ProcessCommitTx(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx) (string, error) {
	nativeAccount := preCheckedTx.From.String()
	txIdStr := preCheckedTx.TxId.String()
	prefix := fmt.Sprintf("%s ProcessCommitTx account=%s txId=%s -", l.prefix, nativeAccount, txIdStr)

	// 1. 分布式锁，防止同一地址并发提交
	mutex := l.rs.NewMutex("stake:reward:commit-tx:"+nativeAccount, redsync.WithExpiry(time.Hour))
	if err := mutex.Lock(); err != nil {
		var errTaken *redsync.ErrTaken
		if !errors.As(err, &errTaken) {
			log.Errorf("%s 获取分布式锁错误: %v", prefix, err)
			return "", errors.New("process transaction error")
		}
		return "", errors.New("you have unfinished service, please try again later")
	}
	defer mutex.Unlock()

	// 2. 服务端重新计算 txInfo（不信任客户端参数）
	txInfo, err := l.GetTxInfo(nativeAccount)
	if err != nil {
		log.Errorf("%s 获取交易信息错误: %v", prefix, err)
		return "", fmt.Errorf("get tx info error: %v", err)
	}

	// 3. 再次查询本次可领取奖励（用于 recordStakeClaim）
	var rewards []model.StakeReward
	if err := l.db.Table(model.TableNameStakeReward).
		Where("native_account = ? AND reward_state = ? AND pending = ?", nativeAccount, 0, false).
		Find(&rewards).Error; err != nil {
		log.Errorf("%s 查询奖励记录错误: %v", prefix, err)
		return "", errors.New("query stake rewards error")
	}

	// 4. 解析交易
	decodedTx, err := l.decodeSOLTx(&preCheckedTx.SOLTx, txInfo)
	if err != nil {
		log.Errorf("%s 解析solana交易错误: %v", prefix, err)
		return "", fmt.Errorf("decode transaction error: %v", err)
	}

	// 5. 校验交易
	if err := l.checkSOLTx(decodedTx, txInfo); err != nil {
		log.Errorf("%s 校验交易错误: %v", prefix, err)
		return "", fmt.Errorf("check transaction error: %v", err)
	}

	// 6. 发送给 base 模块签名 + 广播
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Stake", "StakeReward")
	if err != nil {
		log.Errorf("%s 发送交易失败: %v", prefix, err)
		return "", errors.New(utils.FilterAndTranslateSOLError(err))
	}
	if sentTxId == "" || sentTxId != txIdStr {
		log.Errorf("%s base返回的txId[%s]与预期[%s]不一致", prefix, sentTxId, txIdStr)
		return "", errors.New("sent transaction id not equal expected")
	}

	// 7. 落库：标记 pending + 写 claim record + 写 service_tx
	if err := l.recordStakeClaim(nativeAccount, txIdStr, rewards); err != nil {
		log.Errorf("%s 记录领取信息错误: %v", prefix, err)
		return "", errors.New("record stake claim error")
	}

	log.Infof("%s 处理完成", prefix)
	return txIdStr, nil
}
