package rpc

import (
	"context"
	"encoding/base64"
	"dubbo.apache.org/dubbo-go/v3"
	"dubbo.apache.org/dubbo-go/v3/logger"
	"dubbo.apache.org/dubbo-go/v3/registry"
	"errors"
	"fmt"
	"github.com/gagliardetto/solana-go"
	basepb "oshit-go/app/pb/base"
)

// BaseClient 封装 base 模块的 Dubbo RPC 客户端
type BaseClient struct {
	svc basepb.BaseService
}

// NewBaseClient 初始化 Dubbo Triple 客户端，连接到 base 模块
//
// nacosAddr 格式: "127.0.0.1:8848"
// serviceName 对应 base 模块 dubbo.name 配置，例如 "base-rpc"
func NewBaseClient(nacosAddr, serviceName string) (*BaseClient, error) {
	ins, err := dubbo.NewInstance(
		dubbo.WithName(serviceName),
		dubbo.WithRegistry(
			registry.WithNacos(),
			registry.WithAddress(nacosAddr),
		),
		dubbo.WithLogger(logger.WithLevel("warn")),
	)
	if err != nil {
		return nil, fmt.Errorf("create dubbo instance failed: %w", err)
	}

	cli, err := ins.NewClient()
	if err != nil {
		return nil, fmt.Errorf("create dubbo client failed: %w", err)
	}

	svc, err := basepb.NewBaseService(cli)
	if err != nil {
		return nil, fmt.Errorf("create BaseService client failed: %w", err)
	}

	return &BaseClient{svc: svc}, nil
}

// ---- Fee ----

func (c *BaseClient) GetFeeTolerance(ctx context.Context) (*basepb.GetFeeToleranceRsp, error) {
	return c.svc.GetFeeTolerance(ctx, &basepb.GetFeeToleranceReq{})
}

func (c *BaseClient) GetPriorityFee(ctx context.Context) (*basepb.GetPriorityFeeRsp, error) {
	return c.svc.GetPriorityFee(ctx, &basepb.GetPriorityFeeReq{})
}

func (c *BaseClient) GetInstUnits(ctx context.Context) (*basepb.GetInstUnitsRsp, error) {
	return c.svc.GetInstUnits(ctx, &basepb.GetInstUnitsReq{})
}

// ---- Price ----

func (c *BaseClient) GetTokenQuoteSOLPrice(ctx context.Context) (float64, error) {
	rsp, err := c.svc.GetTokenQuoteSOLPrice(ctx, &basepb.GetTokenQuoteSOLPriceReq{})
	if err != nil {
		return 0, err
	}
	price := rsp.Price
	if price <= 0 {
		return 0, errors.New("got price less or equal than 0")
	}
	return price, nil
}

func (c *BaseClient) GetTokenQuoteUSDTPrice(ctx context.Context) (float64, error) {
	rsp, err := c.svc.GetTokenQuoteUSDTPrice(ctx, &basepb.GetTokenQuoteUSDTPriceReq{})
	if err != nil {
		return 0, err
	}
	price := rsp.Price
	if price <= 0 {
		return 0, errors.New("got price less or equal than 0")
	}
	return price, nil
}

func (c *BaseClient) GetUSDTQuoteSOLPrice(ctx context.Context) (float64, error) {
	rsp, err := c.svc.GetUSDTQuoteSOLPrice(ctx, &basepb.GetUSDTQuoteSOLPriceReq{})
	if err != nil {
		return 0, err
	}
	price := rsp.Price
	if price <= 0 {
		return 0, errors.New("got price less or equal than 0")
	}
	return price, nil
}

// ---- Invite ----

func (c *BaseClient) GetAccountByInviteCode(ctx context.Context, inviteCode string) (*basepb.GetAccountByInviteCodeRsp, error) {
	rsp, err := c.svc.GetAccountByInviteCode(ctx, &basepb.GetAccountByInviteCodeReq{InviteCode: inviteCode})
	if err != nil {
		return nil, err
	}
	// RecordId 为空说明记录不存在（base 侧返回空结构体）
	if rsp == nil || rsp.RecordId == "" {
		return nil, nil
	}
	return rsp, nil
}

func (c *BaseClient) CheckInviteRecord(ctx context.Context, nativeAccount string) (bool, error) {
	rsp, err := c.svc.CheckInviteRecord(ctx, &basepb.CheckInviteRecordReq{NativeAccount: nativeAccount})
	if err != nil {
		return false, err
	}
	return rsp.Exists, nil
}

func (c *BaseClient) GetUpInviterRecords(ctx context.Context, nativeAccount string, depth int32) (*basepb.RecursiveQueryRsp, error) {
	return c.svc.GetUpInviterRecords(ctx, &basepb.RecursiveQueryReq{
		NativeAccount: nativeAccount,
		Depth:         depth,
	})
}

func (c *BaseClient) GetDownInviteeRecords(ctx context.Context, nativeAccount string, depth int32) (*basepb.RecursiveQueryRsp, error) {
	return c.svc.GetDownInviteeRecords(ctx, &basepb.RecursiveQueryReq{
		NativeAccount: nativeAccount,
		Depth:         depth,
	})
}

// SendTransaction 将 Solana 交易序列化后发给 base 模块签名并异步广播
// service / subService 用于在 t_service_key 中查找对应私钥
func (c *BaseClient) SendTransaction(ctx context.Context, tx *solana.Transaction, service, subService string) (string, error) {
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("marshal transaction error: %v", err)
	}
	encodedTx := base64.StdEncoding.EncodeToString(txBytes)

	rsp, err := c.svc.SendTransaction(ctx, &basepb.SendTransactionReq{
		EncodedTx:  encodedTx,
		Service:    service,
		SubService: subService,
	})
	if err != nil {
		return "", err
	}
	return rsp.TxId, nil
}

func (c *BaseClient) FindInviteRelationByAccount(ctx context.Context, nativeAccount string) (*basepb.InviteRelation, error) {
	rsp, err := c.svc.FindInviteRelationByAccount(ctx, &basepb.FindInviteRelationByAccountReq{NativeAccount: nativeAccount})
	if err != nil {
		return nil, err
	}
	if !rsp.Found {
		return nil, nil
	}
	return rsp.Record, nil
}
