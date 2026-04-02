package server

import (
	"context"

	"oshit-go/app/base/api/internal/logic"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	basepb "oshit-go/app/pb/base"
	"oshit-go/common/pkg/entity"
)

// BaseRpcService 实现 BaseServiceHandler 接口，委托给 logic 层
type BaseRpcService struct {
	svcCtx *svc.ServiceContext
}

func NewBaseRpcService(svcCtx *svc.ServiceContext) *BaseRpcService {
	return &BaseRpcService{svcCtx: svcCtx}
}

// ---- Fee ----

func (s *BaseRpcService) GetFeeTolerance(ctx context.Context, _ *basepb.GetFeeToleranceReq) (*basepb.GetFeeToleranceRsp, error) {
	rsp, err := logic.NewFeeLogic(ctx, s.svcCtx).GetFeeTolerance()
	if err != nil {
		return nil, err
	}
	return &basepb.GetFeeToleranceRsp{MaxFeeLess: rsp.MaxLessRate}, nil
}

func (s *BaseRpcService) GetPriorityFee(ctx context.Context, _ *basepb.GetPriorityFeeReq) (*basepb.GetPriorityFeeRsp, error) {
	rsp, err := logic.NewFeeLogic(ctx, s.svcCtx).GetPriorityFee()
	if err != nil {
		return nil, err
	}
	return &basepb.GetPriorityFeeRsp{
		PerComputeUnit: feeDetailPb(rsp.PerComputeUnit),
		PerTransaction: feeDetailPb(rsp.PerTransaction),
	}, nil
}

func (s *BaseRpcService) GetInstUnits(ctx context.Context, _ *basepb.GetInstUnitsReq) (*basepb.GetInstUnitsRsp, error) {
	rsp, err := logic.NewFeeLogic(ctx, s.svcCtx).GetInstUnits()
	if err != nil {
		return nil, err
	}
	return &basepb.GetInstUnitsRsp{
		MiniRent:          rsp.MiniRent,
		AssociatedAccount: rsp.AssociatedAccount,
		TransferChecked:   rsp.TransferChecked,
		Memo:              rsp.Memo,
	}, nil
}

// ---- Price ----

func (s *BaseRpcService) GetTokenQuoteSOLPrice(ctx context.Context, _ *basepb.GetTokenQuoteSOLPriceReq) (*basepb.PriceQuoteRsp, error) {
	rsp, err := logic.NewPriceLogic(ctx, s.svcCtx).GetTokenQuoteSOLPrice()
	if err != nil {
		return nil, err
	}
	return &basepb.PriceQuoteRsp{Price: rsp.Price}, nil
}

func (s *BaseRpcService) GetTokenQuoteUSDTPrice(ctx context.Context, _ *basepb.GetTokenQuoteUSDTPriceReq) (*basepb.PriceQuoteRsp, error) {
	rsp, err := logic.NewPriceLogic(ctx, s.svcCtx).GetTokenQuoteUSDTPrice()
	if err != nil {
		return nil, err
	}
	return &basepb.PriceQuoteRsp{Price: rsp.Price}, nil
}

func (s *BaseRpcService) GetUSDTQuoteSOLPrice(ctx context.Context, _ *basepb.GetUSDTQuoteSOLPriceReq) (*basepb.PriceQuoteRsp, error) {
	rsp, err := logic.NewPriceLogic(ctx, s.svcCtx).GetUSDTQuoteSOLPrice()
	if err != nil {
		return nil, err
	}
	return &basepb.PriceQuoteRsp{Price: rsp.Price}, nil
}

// ---- Tx ----

func (s *BaseRpcService) SendTransaction(ctx context.Context, req *basepb.SendTransactionReq) (*basepb.SendTransactionRsp, error) {
	rsp, err := logic.NewTxLogic(ctx, s.svcCtx).SendTransaction(&types.SendTransactionReq{
		EncodedTx:  req.EncodedTx,
		Service:    req.Service,
		SubService: req.SubService,
	})
	if err != nil {
		return nil, err
	}
	return &basepb.SendTransactionRsp{
		RecordId: rsp.RecordID,
		TxId:     rsp.TxID,
	}, nil
}

// ---- 类型转换 ----
func feeDetailPb(f entity.FeeDetail) *basepb.FeeDetail {
	return &basepb.FeeDetail{
		Low:     f.Low,
		Medium:  f.Medium,
		High:    f.High,
		Extreme: f.Extreme,
	}
}
