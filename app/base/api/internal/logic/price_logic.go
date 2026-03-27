package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	"strconv"
)

type PriceLogic struct {
	ctx    context.Context
	srvCtx *svc.ServiceContext
}

func NewPriceLogic(ctx context.Context, srvCtx *svc.ServiceContext) *PriceLogic {
	return &PriceLogic{
		ctx:    ctx,
		srvCtx: srvCtx,
	}
}

// GetTokenQuoteSOLPrice 获取token兑换SOL价格
func (l *PriceLogic) GetTokenQuoteSOLPrice() (*types.PriceQuoteRsp, error) {
	// 计算所有的token的价格
	quoteSOLPriceStr, err := l.srvCtx.Redis.Get(context.Background(), "RAYDIUM-QUOTE-SOL-PRICE").Result()
	if err != nil {
		log.Errorf("计算奖励金额价格错误: %v", err)
		return nil, errors.New("get toke quote sol price failed")
	}
	quoteSOLPrice, err := strconv.ParseFloat(quoteSOLPriceStr, 64)
	if err != nil {
		log.Errorf("从redis当中获取token兑换solana价格错误: %v", err)
		return nil, errors.New("get toke quote sol price failed")
	}
	rawPrice := quoteSOLPrice / float64(solana.LAMPORTS_PER_SOL)
	return &types.PriceQuoteRsp{
		Price: rawPrice,
	}, nil
}

// GetTokenQuoteUSDTPrice 获取token兑换USDT价格
func (l *PriceLogic) GetTokenQuoteUSDTPrice() (*types.PriceQuoteRsp, error) {
	// 计算所有的token的价格
	quoteSOLPriceStr, err := l.srvCtx.Redis.Get(context.Background(), "RAYDIUM-QUOTE-USDT-PRICE").Result()
	if err != nil {
		log.Errorf("计算奖励金额价格错误: %v", err)
		return nil, errors.New("get toke quote usdt price failed")
	}
	quoteSOLPrice, err := strconv.ParseInt(quoteSOLPriceStr, 10, 64)
	if err != nil {
		log.Errorf("从redis当中获取token兑换solana价格错误: %v", err)
		return nil, errors.New("get toke quote usdt price failed")
	}
	rawPrice := float64(quoteSOLPrice) / 1e6
	return &types.PriceQuoteRsp{Price: rawPrice}, nil
}

// GetUSDTQuoteSOLPrice 获取USDT兑换SOL价格
func (l *PriceLogic) GetUSDTQuoteSOLPrice() (*types.PriceQuoteRsp, error) {
	// 计算USDT兑换SOL的价格
	priceStr, err := l.srvCtx.Redis.Get(context.Background(), "RAYDIUM-USDT-QUOTE-SOL-PRICE").Result()
	if err != nil {
		log.Errorf("计算奖励金额价格错误: %v", err)
		return nil, errors.New("get usdt quote sol price failed")
	}
	quoteSOLPrice, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		log.Errorf("从redis当中获取token兑换solana价格错误: %v", err)
		return nil, errors.New("get usdt quote sol price failed")
	}
	return &types.PriceQuoteRsp{
		Price: quoteSOLPrice,
	}, nil
}

// GetBirdEyePrice 获取BirdEye价格数据
func (l *PriceLogic) GetBirdEyePrice(req *types.GetBirdEyePriceReq) (*types.BirdEyePriceRsp, error) {
	// 构建Redis键名
	redisKey := fmt.Sprintf("BIRD-EYE-APP-PRICE-%s", req.Interval)

	// 从Redis获取数据
	var rawData interface{}
	ctx := context.Background()
	val, err := l.srvCtx.Redis.Get(ctx, redisKey).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return nil, errors.New("can not get price data")
	case err != nil:
		log.Errorf("获取bird eye 价格信息 - Redis查询失败 [%s]: %v", redisKey, err)
		return nil, errors.New("server internal error")
	default:
		if err := json.Unmarshal([]byte(val), &rawData); err != nil {
			log.Errorf("获取bird eye 价格信息 - JSON解析失败 [%s]: %v", redisKey, err)
			return nil, errors.New("server internal error")
		}
	}

	return &types.BirdEyePriceRsp{
		Data: rawData,
	}, nil
}
