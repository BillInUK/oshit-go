package logic

import (
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	"oshit-go/common/pkg/entity"
	"strconv"
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

	var perComputeUnit entity.FeeDetail
	if err := json.Unmarshal([]byte(perComputeUnitJSON), &perComputeUnit); err != nil {
		return nil, errors.New("unmarshal per compute unit fee error")
	}

	// 从Redis获取每交易手续费
	perTransactionJSON, err := l.srvCtx.Redis.Get(l.ctx, "SOL-PRIORITY-FEE-PER-TRANSACTION-ON-BLOCKCHAIN").Result()
	if err != nil {
		return nil, errors.New("get per transaction fee error")
	}

	var perTransaction entity.FeeDetail
	if err := json.Unmarshal([]byte(perTransactionJSON), &perTransaction); err != nil {
		return nil, errors.New("unmarshal per transaction fee error")
	}

	return &types.PriorityFeeRsp{
		PerComputeUnit: perComputeUnit,
		PerTransaction: perTransaction,
	}, nil
}

// GetInstUnits 获取计算单元消耗
func (l *FeeLogic) GetInstUnits() (*types.ComputeUnitConsumedRsp, error) {
	ctx := context.Background()
	miniRentStr, err := l.srvCtx.Redis.Get(ctx, "COMPUTE-UNIT-ASSOCIATED-ACCOUNT-MINI-RENT").Result()
	if err != nil {
		log.Errorf("获取solana 最小账户租金错误: %v", err)
		return nil, errors.New("get mini rent error")
	}
	miniRent, err := strconv.ParseUint(miniRentStr, 10, 64)
	if err != nil {
		log.Errorf("获取solana 最小账户租金，无法将redis内的值转成float64错误: %v", err)
		return nil, errors.New("get mini rent error")
	}
	associatedAccountStr, err := l.srvCtx.Redis.Get(ctx, "COMPUTE-UNIT-ASSOCIATED-ACCOUNT").Result()
	if err != nil {
		log.Errorf("获取solana 获取创建token account消耗计算单元错误: %v", err)
		return nil, errors.New("get associated account units consumed error")
	}
	associatedAccount, err := strconv.ParseUint(associatedAccountStr, 10, 64)
	if err != nil {
		log.Errorf("获取solana 获取创建token account消耗计算单元，无法将redis内的值转成float64错误: %v", err)
		return nil, errors.New("get mini rent error")
	}
	memoStr, err := l.srvCtx.Redis.Get(ctx, "COMPUTE-UNIT-MEMO").Result()
	if err != nil {
		log.Errorf("获取solana 获取memo指令消耗计算单元，将redis内的值转成float64错误: %v", err)
		return nil, errors.New("get transfer checked units consumed error")
	}
	memo, err := strconv.ParseUint(memoStr, 10, 64)
	if err != nil {
		log.Errorf("获取solana 获取memo指令消耗计算单元，将redis内的值转成float64错误: %v", err)
		return nil, errors.New("get mini rent error")
	}
	return &types.ComputeUnitConsumedRsp{
		MiniRent:          miniRent,
		AssociatedAccount: associatedAccount,
		TransferChecked:   6472,
		Memo:              memo,
	}, nil
}
