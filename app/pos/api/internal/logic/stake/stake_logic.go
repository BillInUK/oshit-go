package stake

import (
	"context"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	posrpc "oshit-go/app/pos/api/internal/rpc"
	"oshit-go/app/pos/api/internal/svc"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/utils"
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
	distLevel     int32                               // 向上递归奖励的层级
	fixConfig     map[int32]model.StakeFixRateConfig  // 每日固定利息配置
	serviceConfig *model.StakeRewardConfig            // 下发奖励配置
	starWhiteList map[string]model.StakeStarWhitelist // 星级用户白名单
	starLevelRule map[int32]model.StakeStarLevelRule  // 星级评定规则

	// 快照逻辑
	snapShotLogic *StakeSnapShotLogic
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
		distLevel:         srvCtx.StakeDistLevel,
		fixConfig:         srvCtx.StakeFixConfig,
		starWhiteList:     srvCtx.StakeStarWhitelist,
		starLevelRule:     srvCtx.StakeStarLevelRule,
		decimals:          uint8(srvCtx.TokenConfig.Decimals),
		LightHouseAddress: srvCtx.LightHouseAddress,
		snapShotLogic:     NewStakeSnapShotLogic(ctx, srvCtx),
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
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Stake", "StakeToken")
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
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Stake", "UnStakeToken")
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
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Stake", "ReStakeToken")
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
