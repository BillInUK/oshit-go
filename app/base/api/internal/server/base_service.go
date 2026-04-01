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

// ---- Invite ----

func (s *BaseRpcService) GetAccountByInviteCode(ctx context.Context, req *basepb.GetAccountByInviteCodeReq) (*basepb.GetAccountByInviteCodeRsp, error) {
	rsp, err := logic.NewInviteLogic(ctx, s.svcCtx).GetAccountByInviteCode(&types.GetAccountByInviteCodeReq{
		InviteCode: req.InviteCode,
	})
	if err != nil {
		return nil, err
	}
	if rsp == nil {
		return &basepb.GetAccountByInviteCodeRsp{}, nil
	}
	return &basepb.GetAccountByInviteCodeRsp{
		RecordId:      rsp.RecordID,
		NativeAccount: rsp.NativeAccount,
		TokenAccount:  rsp.TokenAccount,
		InviteCode:    rsp.InviteCode,
		CreatedAt:     rsp.CreatedAt,
	}, nil
}

func (s *BaseRpcService) FindInviteRelationByAccount(ctx context.Context, req *basepb.FindInviteRelationByAccountReq) (*basepb.FindInviteRelationByAccountRsp, error) {
	record, err := logic.NewInviteLogic(ctx, s.svcCtx).FindInviteRelationByAccount(&types.FindInviteRelationByAccountReq{
		NativeAccount: req.NativeAccount,
	})
	if err != nil {
		return nil, err
	}
	if record == nil {
		return &basepb.FindInviteRelationByAccountRsp{Found: false}, nil
	}
	return &basepb.FindInviteRelationByAccountRsp{
		Found:  true,
		Record: inviteRelationPb(*record),
	}, nil
}

func (s *BaseRpcService) CheckInviteRecord(ctx context.Context, req *basepb.CheckInviteRecordReq) (*basepb.CheckInviteRecordRsp, error) {
	rsp, err := logic.NewInviteLogic(ctx, s.svcCtx).CheckInviteRecord(&types.CheckInviteRecordReq{
		NativeAccount: req.NativeAccount,
	})
	if err != nil {
		return nil, err
	}
	return &basepb.CheckInviteRecordRsp{Exists: rsp.Exists}, nil
}

func (s *BaseRpcService) GetUpInviterRecords(ctx context.Context, req *basepb.RecursiveQueryReq) (*basepb.RecursiveQueryRsp, error) {
	rsp, err := logic.NewInviteLogic(ctx, s.svcCtx).GetUpInviterRecords(&types.RecursiveQueryReq{
		NativeAccount: req.NativeAccount,
		Depth:         int(req.Depth),
	})
	if err != nil {
		return nil, err
	}
	return inviteRelationsPb(rsp), nil
}

func (s *BaseRpcService) GetDownInviteeRecords(ctx context.Context, req *basepb.RecursiveQueryReq) (*basepb.RecursiveQueryRsp, error) {
	rsp, err := logic.NewInviteLogic(ctx, s.svcCtx).GetDownInviteeRecords(&types.RecursiveQueryReq{
		NativeAccount: req.NativeAccount,
		Depth:         int(req.Depth),
	})
	if err != nil {
		return nil, err
	}
	return inviteRelationsPb(rsp), nil
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

func inviteRelationPb(r types.InviteRelation) *basepb.InviteRelation {
	return &basepb.InviteRelation{
		RecordId:             r.RecordID,
		InviterNativeAccount: r.InviterNativeAccount,
		InviterTokenAccount:  r.InviterTokenAccount,
		InviteeNativeAccount: r.InviteeNativeAccount,
		InviteeTokenAccount:  r.InviteeTokenAccount,
		Channel:              r.Channel,
		Level:                r.Level,
		TxId:                 r.TxID,
		CreatedAt:            r.CreatedAt,
	}
}

func inviteRelationsPb(rsp *types.RecursiveQueryRsp) *basepb.RecursiveQueryRsp {
	records := make([]*basepb.InviteRelation, 0, len(rsp.Records))
	for _, r := range rsp.Records {
		records = append(records, inviteRelationPb(r))
	}
	return &basepb.RecursiveQueryRsp{Records: records}
}
