package pos

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

type PosRewardLogic struct {
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
	snapShotLogic *PosSnapShotLogic
}

func NewPosRewardLogic(ctx context.Context, srvCtx *svc.ServiceContext) *PosRewardLogic {
	return &PosRewardLogic{
		prefix:            "Pos业务 -",
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
		serviceConfig: srvCtx.PosRewardConfig,
		snapShotLogic: NewPosSnapShotLogic(ctx, srvCtx),
	}
}

// GetConfig 获取奖励规则配置
func (l *PosRewardLogic) GetConfig() (*model.PosRewardConfig, error) {
	return l.serviceConfig, nil
}

// GetRewardStat 查看pos奖励统计信息
func (l *PosRewardLogic) GetRewardStat(nativeAccount string) (*types.PosRewardDetail, error) {
	var detail types.PosRewardDetail

	// 查询快照当中的持币量
	var snapShot model.PosSnapShot
	table := l.db.Table(model.TableNamePosSnapShot)
	if err := table.Where("native_account = ?", nativeAccount).Order("snap_day DESC").First(&snapShot).Error; err != nil {
		// 如果没有快照则所有结果都为0
		if errors.Is(err, gorm.ErrRecordNotFound) {
			detail.StarLevel = 0
			detail.HoldingAmount = 0
			detail.TeamHoldingAmount = 0
			detail.RewardAmount = 0
			detail.TeamRewardAmount = 0
			return &detail, nil
		}
		log.Errorf("pos业务 - 根据地址 %s 日期 %v 查询快照信息错误: %v", nativeAccount, snapShot.SnapDay, err)
		return nil, err
	}
	detail.HoldingAmount = snapShot.Amount

	log.Infof("pos业务 - 地址 %s 数据库当中的持币金额 %f 持币金额 %d", nativeAccount, snapShot.Amount, detail.HoldingAmount)

	// 查询奖励当中的非星级部分
	var reward model.PosReward
	table = l.db.Table(model.TableNamePosReward)
	if err := table.Where("native_account= ? AND snap_day = CAST(? AS DATE)  AND starred = false", nativeAccount, snapShot.SnapDay).First(&reward).Error; err != nil {
		detail.StarLevel = 0
		detail.RewardAmount = 0
	}
	detail.RewardAmount = reward.RewardAmount
	detail.RewardState = reward.RewardState
	detail.Pending = reward.Pending

	// 查询奖励当中的星级部分
	var starredReward model.PosReward
	table = l.db.Table(model.TableNamePosReward)
	if err := table.Where("native_account= ? AND snap_day= CAST(? AS DATE)  AND starred = true", nativeAccount, snapShot.SnapDay).First(&starredReward).Error; err != nil {
		detail.StarLevel = 0
		detail.StarredRewardAmount = 0
	}
	detail.StarLevel = starredReward.StarLevel
	detail.StarredRewardAmount = starredReward.RewardAmount

	// 查询下属团队持币量(不包含自己的持币量)
	dbTeamHoldAmount, err := l.snapShotLogic.GetPosGroupHoldAmount(nativeAccount, snapShot.SnapDay)
	if err != nil {
		log.Errorf("pos业务 - 根据地址 %s 获取团队总持币数错误: %v", nativeAccount, err)
		return nil, err
	}
	detail.TeamHoldingAmount = dbTeamHoldAmount

	// 查询下属团队获取的总奖励数量
	dbTeamRewardAmount, err := l.snapShotLogic.GetGroupTotalFixReward(nativeAccount, snapShot.SnapDay)
	if err != nil {
		log.Errorf("pos业务 - 根据地址 %s 获取团队总奖励数错误: %v", nativeAccount, err)
		return nil, err
	}
	detail.TeamRewardAmount = dbTeamRewardAmount
	detail.TotalRewardAmount = detail.RewardAmount

	// 计算星级奖励当中属于自己的部分和团队带来的部分
	if detail.StarLevel > 0 {
		// 查询下属团队获取的总奖励数量
		dbTeamRewardAmount, err = l.snapShotLogic.GetGroupTotalFixReward(nativeAccount, snapShot.SnapDay)
		if err != nil {
			log.Errorf("pos业务 - 根据地址 %s 获取团队总奖励数错误: %v", nativeAccount, err)
			return nil, err
		}
		detail.TeamRewardAmount = dbTeamRewardAmount

		// 查询星级用户的固定奖励的费率
		var missionConf model.PosMissionConfig
		table = l.db.Table(model.TableNamePosMissionConfig)
		if err := table.Where("reward_type = ? AND starred = ?", types.PosFixedIncome, true).First(&missionConf).Error; err != nil {
			log.Errorf("pos业务 - 查询星级用户每日固定奖励费率错误:%v", err)
			return nil, err
		}

		// 计算星级奖励当中属于自己的部分
		starredRewardBaseF := snapShot.Amount * float64(detail.StarLevel) / 10 * missionConf.Rate / 100 / 365
		detail.StarredRewardBase = starredRewardBaseF
		// 计算团队的给用户带来的星级奖励
		detail.StarredRewardFromTeam = detail.StarredRewardAmount - detail.StarredRewardBase
		detail.TotalRewardAmount += detail.StarredRewardAmount
	}

	return &detail, nil
}

// GetRewards 查看pos发放奖励
func (l *PosRewardLogic) GetRewards(nativeAccount string) ([]model.PosReward, error) {
	var rewards []model.PosReward
	table := l.db.Table(model.TableNamePosReward)
	err := table.Where("native_account = ? and reward_state = ? AND pending = ?", nativeAccount, 0, false).Find(&rewards).Error
	if err != nil {
		return nil, err
	}
	// 返回结果和总数
	return rewards, nil
}

// CalculatePosClaimRewardFee 计算质押费用
func (l *PosRewardLogic) calCostFee(totalRewardAmount, tokenQuoteUSDTPrice, usdtQuoteSOLPrice float64) float64 {
	defaultUSDTFee := totalRewardAmount / 1000 * tokenQuoteUSDTPrice * 0.2
	// 如果小于0.01美金
	if defaultUSDTFee < 0.01 {
		return 0.01 * usdtQuoteSOLPrice * float64(solana.LAMPORTS_PER_SOL)
	}
	// 如果大于1美金
	if defaultUSDTFee > 1 {
		return 1 * usdtQuoteSOLPrice * float64(solana.LAMPORTS_PER_SOL)
	}
	return defaultUSDTFee * usdtQuoteSOLPrice * float64(solana.LAMPORTS_PER_SOL)
}

// GetTxInfo 获取领取pos奖励交易信息
func (l *PosRewardLogic) GetTxInfo(nativeAccount string) (*types.ClaimPosRewardTxInfo, error) {
	var txInfo types.ClaimPosRewardTxInfo

	// 查询当天的所有质押奖励
	totalRewardAmount := 0.0
	if err := l.db.Table(model.TableNamePosReward).
		Select(`COALESCE(SUM(reward_amount), 0) AS reward_amount`).
		Where(`native_account= ? AND reward_state = ? AND pending = ? AND snap_day = CAST(? AS DATE)`, nativeAccount, 0, false, time.Now()).
		Scan(&totalRewardAmount).Error; err != nil {
		log.Errorf("%s 获取POS奖励交易信息 - 查询地址: %s 当日所有质押奖励错误: %v", l.prefix, nativeAccount, err)
		return nil, err
	}

	tokenQuoteUSDTPrice, err := l.baseClient.GetTokenQuoteUSDTPrice(l.ctx)
	if err != nil {
		log.Errorf("%s 获取POS奖励交易信息 - 获取Token兑换USDT费率错误: %v", l.prefix, err)
		return nil, err
	}
	usdtQuoteSOLPrice, err := l.baseClient.GetUSDTQuoteSOLPrice(l.ctx)
	if err != nil {
		log.Errorf("%s 获取POS奖励交易信息 - 获取USDT兑换SOL费率错误: %v", l.prefix, err)
		return nil, err
	}
	costFee := l.calCostFee(totalRewardAmount, tokenQuoteUSDTPrice, usdtQuoteSOLPrice)
	log.Infof("totalRewardAmount %f,tokenQuoteUSDTPrice %f,usdtQuoteSOLPrice %f costFee %f",
		totalRewardAmount, tokenQuoteUSDTPrice, usdtQuoteSOLPrice, costFee)

	txInfo.RewardAccount = l.serviceConfig.RewardAccount
	txInfo.CostAccount = l.serviceConfig.CostAccount
	txInfo.Mint = l.srvCtx.TokenConfig.Mint
	txInfo.TotalReward = totalRewardAmount
	txInfo.Decimals = l.srvCtx.TokenConfig.Decimals
	txInfo.CostFee = costFee

	return &txInfo, nil
}

// ProcessCommitTx 处理前端提交的 stake 奖励领取交易
func (l *PosRewardLogic) ProcessCommitTx(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx) (string, error) {
	nativeAccount := preCheckedTx.From.String()
	txIdStr := preCheckedTx.TxId.String()
	prefix := fmt.Sprintf("%s ProcessCommitTx account=%s txId=%s -", l.prefix, nativeAccount, txIdStr)

	// 1. 分布式锁，防止同一地址并发提交
	mutex := l.rs.NewMutex("pos:reward:commit-tx:"+nativeAccount, redsync.WithExpiry(time.Hour))
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
	var rewards []model.PosReward
	if err := l.db.Table(model.TableNamePosReward).
		Where("native_account = ? AND reward_state = ? AND pending = ?", nativeAccount, 0, false).
		Find(&rewards).Error; err != nil {
		log.Errorf("%s 查询奖励记录错误: %v", prefix, err)
		return "", errors.New("query pos rewards error")
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
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Pos", "PosReward")
	if err != nil {
		log.Errorf("%s 发送交易失败: %v", prefix, err)
		return "", errors.New(utils.FilterAndTranslateSOLError(err))
	}
	if sentTxId == "" || sentTxId != txIdStr {
		log.Errorf("%s base返回的txId[%s]与预期[%s]不一致", prefix, sentTxId, txIdStr)
		return "", errors.New("sent transaction id not equal expected")
	}

	// 7. 落库：标记 pending + 写 claim record + 写 service_tx
	if err := l.recordPosClaim(nativeAccount, txIdStr, rewards); err != nil {
		log.Errorf("%s 记录领取信息错误: %v", prefix, err)
		return "", errors.New("record stake claim error")
	}

	log.Infof("%s 处理完成", prefix)
	return txIdStr, nil
}

// recordPosClaim 在同一数据库事务中：
// 1. 将本次可领取奖励标记为 pending=true
// 2. 写入 t_service_tx
// 3. 写入 t_pos_reward_claim_record
func (l *PosRewardLogic) recordPosClaim(nativeAccount, txId string, rewards []model.PosReward) error {
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
	if err := dbTx.Table(model.TableNamePosReward).
		Where("native_account = ? AND reward_state = ? AND pending = ?", nativeAccount, 0, false).
		Updates(map[string]interface{}{
			"pending":    true,
			"tx_id":      txId,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("mark stake rewards as pending error: %v", err)
	}

	// 2. 写入 t_stake_reward_claim_record
	claimRecord := model.PosRewardClaimRecord{
		RewardIds: rewardIdsStr,
		TxID:      txId,
		TxState:   0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := dbTx.Table(model.TableNameStakeRewardClaim).Omit("record_id").Create(&claimRecord).Error; err != nil {
		return fmt.Errorf("insert pos reward claim record error: %v", err)
	}

	if err := dbTx.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction error: %v", err)
	}
	return nil
}
