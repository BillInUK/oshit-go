package logic

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
)

type FeeLogic struct {
	ctx    context.Context
	srvCtx *svc.ServiceContext
}

func NewFeeLogic(ctx context.Context, srvCtx *svc.ServiceContext) *FeeLogic {
	return &FeeLogic{
		ctx:    ctx,
		srvCtx: srvCtx,
	}
}

// GetPriorityFee 获取优先手续费
func (l *FeeLogic) GetPriorityFee() (*types.PriorityFeeRsp, error) {
	// 从Redis获取每计算单元手续费
	perComputeUnitJSON, err := l.srvCtx.Redis.Get(l.ctx, "SOL-PRIORITY-FEE-PER-COMPUTE-UNIT-ON-BLOCKCHAIN").Result()
	if err != nil {
		return nil, errors.New("get per compute unit fee error")
	}

	var perComputeUnit types.FeeDetail
	if err := json.Unmarshal([]byte(perComputeUnitJSON), &perComputeUnit); err != nil {
		return nil, errors.New("unmarshal per compute unit fee error")
	}

	// 从Redis获取每交易手续费
	perTransactionJSON, err := l.srvCtx.Redis.Get(l.ctx, "SOL-PRIORITY-FEE-PER-TRANSACTION-ON-BLOCKCHAIN").Result()
	if err != nil {
		return nil, errors.New("get per transaction fee error")
	}

	var perTransaction types.FeeDetail
	if err := json.Unmarshal([]byte(perTransactionJSON), &perTransaction); err != nil {
		return nil, errors.New("unmarshal per transaction fee error")
	}

	return &types.PriorityFeeRsp{
		PerComputeUnit: perComputeUnit,
		PerTransaction: perTransaction,
	}, nil
}

// GetPriorityFeeOnBlockchain 获取链上优先手续费
func (l *FeeLogic) GetPriorityFeeOnBlockchain() (*types.PriorityFeeRsp, error) {
	// 与GetPriorityFee相同，因为数据都来自链上
	return l.GetPriorityFee()
}

// GetComputeUnitConsumed 获取计算单元消耗
func (l *FeeLogic) GetComputeUnitConsumed() (*types.ComputeUnitConsumedRsp, error) {
	// 通过TaskManager获取计算单元消耗
	if l.srvCtx.TaskManager == nil {
		return nil, errors.New("task manager not initialized")
	}

	computeUnitTask := l.srvCtx.TaskManager.GetComputeUnitTask()
	if computeUnitTask == nil {
		return nil, errors.New("compute unit task not found")
	}

	consumed, err := computeUnitTask.GetComputeUnitConsumed(l.ctx)
	if err != nil {
		return nil, errors.New("get compute unit consumed error")
	}

	return &types.ComputeUnitConsumedRsp{
		MiniRent:          consumed["mini_rent"],
		AssociatedAccount: consumed["associated_account"],
		TransferChecked:   consumed["transfer_checked"],
		Memo:              consumed["memo"],
	}, nil
}
