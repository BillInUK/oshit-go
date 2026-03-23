package logic

import (
	"context"
	"github.com/pkg/errors"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
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
	// 简化实现：从Redis获取价格
	// 实际应该调用价格服务
	_, err := l.srvCtx.Redis.Get(l.ctx, "RAYDIUM-QUOTE-SOL-PRICE").Result()
	if err != nil {
		// 返回默认值
		return &types.PriceQuoteRsp{
			Price: 0.0015,
		}, nil
	}

	// 解析价格
	// 注意：实际实现需要根据存储格式解析
	return &types.PriceQuoteRsp{
		Price: 0.0015, // 简化实现
	}, nil
}

// GetTokenQuoteUSDTPrice 获取token兑换USDT价格
func (l *PriceLogic) GetTokenQuoteUSDTPrice() (*types.PriceQuoteRsp, error) {
	// 简化实现：从Redis获取价格
	_, err := l.srvCtx.Redis.Get(l.ctx, "RAYDIUM-QUOTE-USDT-PRICE").Result()
	if err != nil {
		// 返回默认值
		return &types.PriceQuoteRsp{
			Price: 0.0012,
		}, nil
	}

	// 解析价格
	return &types.PriceQuoteRsp{
		Price: 0.0012, // 简化实现
	}, nil
}

// GetUSDTQuoteSOLPrice 获取USDT兑换SOL价格
func (l *PriceLogic) GetUSDTQuoteSOLPrice() (*types.PriceQuoteRsp, error) {
	// 简化实现：从Redis获取价格
	_, err := l.srvCtx.Redis.Get(l.ctx, "RAYDIUM-USDT-QUOTE-SOL-PRICE").Result()
	if err != nil {
		// 返回默认值
		return &types.PriceQuoteRsp{
			Price: 1.25,
		}, nil
	}

	// 解析价格
	return &types.PriceQuoteRsp{
		Price: 1.25, // 简化实现
	}, nil
}

// GetBirdEyePrice 获取BirdEye价格数据
func (l *PriceLogic) GetBirdEyePrice(req *types.GetBirdEyePriceReq) (*types.BirdEyePriceRsp, error) {
	// 验证时间间隔
	validIntervals := map[string]bool{"1D": true, "1W": true, "1M": true}
	if !validIntervals[req.Interval] {
		return nil, errors.New("invalid interval, must be 1D, 1W, or 1M")
	}

	// 通过TaskManager获取价格数据
	if l.srvCtx.TaskManager == nil {
		return nil, errors.New("task manager not initialized")
	}

	priceTask := l.srvCtx.TaskManager.GetPriceTask()
	if priceTask == nil {
		return nil, errors.New("price task not found")
	}

	priceData, err := priceTask.GetBirdEyePrice(l.ctx, req.Interval)
	if err != nil {
		return nil, errors.New("get bird eye price error")
	}

	return &types.BirdEyePriceRsp{
		Data: priceData,
	}, nil
}
