package campaign

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
	"oshit-go/common/pkg/dal/query"
	"oshit-go/common/utils"
	"strconv"
	"time"
)

const (
	ExchangeSys         = "oshit reward"
	ExchangeBiz         = "exchange score to token"
	ReasonFreeze        = "freeze before tx confirmed"
	ReasonUnFreeze      = "transaction failed"
	ReasonConsumeFreeze = "transaction success"
)

type CampaignLogic struct {
	prefix        string
	ctx           context.Context
	srvCtx        *svc.ServiceContext
	db            *gorm.DB
	rd            redis.UniversalClient
	rs            redsync.Redsync
	rpcClient     *rpc.Client
	baseClient    *rewardrpc.BaseClient
	serviceConfig *model.CampaignExchangeConfig
}

func NewCampaignLogic(ctx context.Context, srvCtx *svc.ServiceContext) *CampaignLogic {
	return &CampaignLogic{
		prefix:        "Campaign业务 -",
		ctx:           ctx,
		srvCtx:        srvCtx,
		db:            srvCtx.DB,
		rd:            srvCtx.Redis,
		rs:            srvCtx.RedSync,
		rpcClient:     srvCtx.RpcClient,
		baseClient:    srvCtx.BaseClient,
		serviceConfig: srvCtx.CampaignExchangeConfig,
	}
}

func (l *CampaignLogic) GetExchangeQuotaInfo(ctx context.Context, userId uint64) (*model.UserDailyExchangeQuota, *model.GlobalDailyExchangeLimit, error) {
	prefix := fmt.Sprintf("%s 查询当天额度信息 - 用户id: %d", l.prefix, userId)

	// 1. 查询当天的全局token兑换额度
	currentDate := time.Now().UTC().Truncate(24 * time.Hour)
	q := query.Use(l.db)
	globalLimitDo := q.GlobalDailyExchangeLimit.WithContext(ctx)
	globalLimit, err := globalLimitDo.Where(q.GlobalDailyExchangeLimit.QuotaDate.Eq(currentDate)).First()
	if err != nil {
		// 如果当天没有记录，创建一条默认记录
		if errors.Is(err, gorm.ErrRecordNotFound) {
			globalLimit = &model.GlobalDailyExchangeLimit{
				DailyLimit: 50000000,
				QuotaDate:  currentDate,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			if err = globalLimitDo.Create(globalLimit); err != nil {

			}
		} else {
			log.Errorf("%s 查询全局兑换限额错误: %v", prefix, err)
			return nil, nil, fmt.Errorf("query global daily limit error")
		}
	}

	// 2. 查询用户的兑换额度是否足够
	userIdStr := strconv.FormatUint(userId, 10)
	userQuotaDo := q.UserDailyExchangeQuota.WithContext(ctx)
	userQuota, err := userQuotaDo.Where(
		q.UserDailyExchangeQuota.UserID.Eq(userIdStr),
		q.UserDailyExchangeQuota.QuotaDate.Eq(currentDate),
	).First()
	if err != nil {
		// 如果用户当天没有记录，创建一条默认记录
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userQuota = &model.UserDailyExchangeQuota{
				UserID:         userIdStr,
				QuotaDate:      currentDate,
				MaxQuota:       50000000,
				FrozenQuota:    0,
				AvailableQuota: 50000000,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			if err = userQuotaDo.Create(userQuota); err != nil {

			}
		} else {
			log.Errorf("%s 查询用户兑换额度错误: %v", prefix, err)
			return nil, nil, fmt.Errorf("query user daily quota error")
		}
	}

	return userQuota, globalLimit, nil
}

// GetTxInfo 获取打包交易信息
func (l *CampaignLogic) GetTxInfo(ctx context.Context, userId uint64, scoreUIAmount uint64) (*types.CampaignExchangeTxInfo, error) {
	prefix := fmt.Sprintf("%s 获取兑换交易信息 - 用户id %d 兑换积分额度 %d", l.prefix, userId, scoreUIAmount)

	// 计算本次兑换的token数量（单位：1/1000 token）
	tokenUIAmount := float64(scoreUIAmount) * l.serviceConfig.Rate / 100
	tokenRawAmount := tokenUIAmount * l.srvCtx.TokenDecimal

	// 计算成本费
	costRate := l.serviceConfig.CostRate / 100
	quoteSOLPrice, err := l.srvCtx.BaseClient.GetTokenQuoteSOLPrice(l.ctx)
	if err != nil {
		log.Errorf("%s 解析交易错误，无法查询到token兑换solana的费率，错误:%v", prefix, err)
		return nil, errors.New("can not get quote sol price")
	}
	costRawFee := tokenUIAmount * quoteSOLPrice * costRate * float64(solana.LAMPORTS_PER_SOL)

	// 返回最终结构体
	return &types.CampaignExchangeTxInfo{
		RewardAccount: l.serviceConfig.RewardAccount,
		Mint:          l.srvCtx.TokenConfig.Mint,
		Decimals:      l.srvCtx.TokenConfig.Decimals,
		CostAccount:   l.serviceConfig.CostAccount,
		TokenAmount:   tokenRawAmount,
		CostFee:       costRawFee,
	}, nil
}

//func (l *CampaignLogic) ProcessCommitTx(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx, req types.CampaignExchangeReq, userId uint64) error {
//	txIdStr := preCheckedTx.TxId.String()
//	scoreUIAmount := req.Score
//	prefix := fmt.Sprintf("%s 处理用户提交交易Id %v -", l.prefix, txIdStr)
//
//	// 1. 解析出来交易里面的 transfer checked 和 transfer 指令集合
//	decodedTx, err := app_utils.DecodeSolanaTransaction(l.rpcClient, l.db, &preCheckedTx.SOLTx, solana.Signature{})
//	if err != nil {
//		log.Errorf("%s 解析solana交易错误: %v", prefix, err)
//		return fmt.Errorf("decode transaction error:%s", err)
//	}
//
//	// 2. 检查transfer checked 指令和transfer 指令是否符合奖励要求
//	decodedServiceTx, err := l.checkSOLTx(scoreUIAmount, decodedTx, preCheckedTx.From.String())
//	if err != nil {
//		log.Errorf("%s 校验solana交易当中的指令错误: %v", prefix, err)
//		return fmt.Errorf("check transaction instruction failed: %v", err)
//	}
//
//	// 3. 查询当天的全局token兑换额度
//	userQuota, globalLimit, err := l.GetExchangeQuotaInfo(ctx, userId)
//	if err != nil {
//		log.Errorf("%s 查询当天额度信息错误: %v", prefix, err)
//		return fmt.Errorf("get exchange quota info error: %v", err)
//	}
//
//	// 计算本次兑换的token数量（单位：1/1000 token）
//	tokenUIAmount := float64(scoreUIAmount) * l.serviceConfig.Rate
//	tokenInstAmount := uint64(tokenUIAmount * l.srvCtx.TokenDecimal)
//
//	// 检查全局额度是否足够
//	if float64(tokenInstAmount) > globalLimit.DailyLimit {
//		log.Errorf("%s 全局兑换额度不足，本次兑换需要 %f，剩余额度 %f", prefix, float64(tokenInstAmount), globalLimit.DailyLimit)
//		return fmt.Errorf("insufficient global daily exchange limit")
//	}
//
//	// 检查用户额度是否足够
//	if float64(tokenInstAmount) > userQuota.AvailableQuota {
//		log.Errorf("%s 用户兑换额度不足，本次兑换需要 %f，剩余可用额度 %f", prefix, float64(tokenInstAmount), userQuota.AvailableQuota)
//		return fmt.Errorf("insufficient user daily exchange quota")
//	}
//
//	// 查询积分
//	scoreData, err := l.srvCtx.CampaignClientV1.QueryScore(userId)
//	if err != nil {
//		log.Errorf("%s 校验solana交易当中的指令错误: %v", prefix, err)
//		return fmt.Errorf("query user campaign score amount failed")
//	}
//	if scoreData.TotalAvailable < scoreUIAmount*1000 {
//		log.Errorf("%s 请求兑换积分 %d 大于可兑换积分 %d", prefix, scoreUIAmount*1000, scoreData.TotalAvailable)
//		return fmt.Errorf("insufficient available score amount")
//	}
//	// 4. 从交易签名获取txId（Solana交易的第一个签名就是交易ID）
//	txId := preCheckedTx.SOLTx.Signatures[0]
//
//	// 5. 请求冻结积分
//	opResult, err := l.srvCtx.CampaignClientV1.FreezeScore(userId, txId.String(), scoreUIAmount*1000, ExchangeSys, ExchangeBiz, ReasonFreeze)
//	if err != nil {
//		log.Errorf("%s 用户 %s 请求冻结积分错误: %v", prefix, userQuota.UserID, err)
//		return fmt.Errorf("freeze core error")
//	}
//
//	// 6. 通过 base 模块的dubbo接口签名并异步广播
//	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Campaign", "ExchangeToken")
//	if err != nil {
//		log.Errorf("%s 调用base模块dubbo接口发送交易失败,错误: %v", prefix, err)
//		return errors.New(utils.FilterAndTranslateSOLError(err))
//	}
//	// 如果发送的交易Id与预期的不一致，则报错
//	if sentTxId == "" || sentTxId != txIdStr {
//		log.Errorf("%s 调用base模块dubbo接口发送的交易Id %s 与预期的不一致", prefix, sentTxId)
//		return errors.New("sent transaction id not equal expected")
//	}
//
//	// 7. 记录领取记录到数据库（包含冻结用户兑换额度）
//	decodedServiceTx.TxID = txId.String()
//	if err = l.recordExchangeRecord(decodedServiceTx, userQuota.UserID, opResult, float64(tokenInstAmount), userQuota, globalLimit); err != nil {
//		log.Errorf("%s 记录交易信息,错误: %v", prefix, err)
//		return errors.New("record official transfer token error")
//	}
//
//	return nil
//}

func (l *CampaignLogic) ProcessCommitTx(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx, req types.CampaignExchangeReq, userId uint64) error {
	userIdStr := strconv.FormatUint(userId, 10)
	txIdStr := preCheckedTx.TxId.String()
	scoreUIAmount := req.Score
	prefix := fmt.Sprintf("%s 处理用户提交交易Id %v -", l.prefix, txIdStr)

	// 分布式锁让同1个用户同时只能兑换1笔积分
	mutex := l.rs.NewMutex("campaign:exchange:process:commit-tx:"+userIdStr, redsync.WithExpiry(1*time.Hour))
	if err := mutex.Lock(); err != nil {
		var errTaken *redsync.ErrTaken
		if !errors.As(err, &errTaken) {
			log.Errorf("%s 获取处理交易的分布式锁错误: %v", prefix, err)
			return errors.New("process transaction error")
		}
		return errors.New("you have unfinished service.please try again later")
	}
	defer mutex.Unlock()

	// 1. 解析出来交易里面的 transfer checked 和 transfer 指令集合
	decodedTx, err := app_utils.DecodeSolanaTransaction(l.rpcClient, l.db, &preCheckedTx.SOLTx, solana.Signature{})
	if err != nil {
		log.Errorf("%s 解析solana交易错误: %v", prefix, err)
		return fmt.Errorf("decode transaction error:%s", err)
	}

	// 2. 获取交易信息
	txInfo, err := l.GetTxInfo(ctx, userId, scoreUIAmount)
	if err != nil {
		log.Errorf("%s 获取交易信息错误: %v", prefix, err)
		return fmt.Errorf("decode transaction error:%s", err)
	}

	// 3. 检查transfer checked 指令和transfer 指令是否符合奖励要求
	decodedServiceTx, err := l.checkSOLTx(txInfo, decodedTx, preCheckedTx.From.String())
	if err != nil {
		log.Errorf("%s 校验solana交易当中的指令错误: %v", prefix, err)
		return fmt.Errorf("check transaction instruction failed: %v", err)
	}

	// 4. 查询当天的全局token兑换额度
	userQuota, globalLimit, err := l.GetExchangeQuotaInfo(ctx, userId)
	if err != nil {
		log.Errorf("%s 查询当天额度信息错误: %v", prefix, err)
		return fmt.Errorf("get exchange quota info error: %v", err)
	}

	// 5. 从交易签名获取txId（Solana交易的第一个签名就是交易ID）
	txId := preCheckedTx.SOLTx.Signatures[0]

	// 6. 请求冻结积分
	opResult, err := l.srvCtx.CampaignClientV1.FreezeScore(userId, txId.String(), scoreUIAmount*1000, ExchangeSys, ExchangeBiz, ReasonFreeze)
	if err != nil {
		log.Errorf("%s 用户 %d 请求冻结积分错误: %v", prefix, userId, err)
		return fmt.Errorf("freeze core error")
	}

	// 7. 通过 base 模块的dubbo接口签名并异步广播
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Campaign", "ExchangeToken")
	if err != nil {
		log.Errorf("%s 调用base模块dubbo接口发送交易失败,错误: %v", prefix, err)
		return errors.New(utils.FilterAndTranslateSOLError(err))
	}
	// 如果发送的交易Id与预期的不一致，则报错
	if sentTxId == "" || sentTxId != txIdStr {
		log.Errorf("%s 调用base模块dubbo接口发送的交易Id %s 与预期的不一致", prefix, sentTxId)
		return errors.New("sent transaction id not equal expected")
	}

	// 8. 记录领取记录到数据库（包含冻结用户兑换额度）
	decodedServiceTx.TxID = txId.String()
	if err = l.recordExchangeRecord(decodedServiceTx, userId, opResult, txInfo.TokenAmount, userQuota, globalLimit); err != nil {
		log.Errorf("%s 记录交易信息,错误: %v", prefix, err)
		return errors.New("record official transfer token error")
	}

	return nil
}
