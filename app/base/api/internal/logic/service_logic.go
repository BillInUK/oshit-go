package logic

import (
	"context"
	"errors"
	"fmt"
	"oshit-go/app/base/api/internal/svc"
	"oshit-go/app/base/api/internal/types"
	"oshit-go/app/base/dal/model"
	"oshit-go/app/base/dal/query"
	"time"

	"gorm.io/gorm"
)

type ServiceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	db     *gorm.DB
}

func NewServiceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ServiceLogic {
	return &ServiceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		db:     svcCtx.DB,
	}
}

// RegisterService 注册服务
func (l *ServiceLogic) RegisterService(req *types.ServiceRegisterReq) (*types.ServiceRegisterRsp, error) {
	q := query.Use(l.db)
	serviceInfo := q.ServiceInfo

	// 检查服务是否已存在
	existing, err := serviceInfo.WithContext(l.ctx).
		Where(serviceInfo.Service.Eq(req.Service)).
		First()

	if err == nil && existing != nil {
		return &types.ServiceRegisterRsp{
			Success: false,
			Message: "service already exists",
		}, nil
	}

	// 验证hook_type配置
	if req.HookType == 0 {
		// Kafka通知，需要mq_topic
		if req.MqTopic == "" {
			return &types.ServiceRegisterRsp{
				Success: false,
				Message: "mq_topic is required when hook_type is 0",
			}, nil
		}
		// Kafka不需要mq_group，可以忽略或用于其他用途
	} else if req.HookType == 1 {
		// Webhook通知，需要webhook
		if req.Webhook == "" {
			return &types.ServiceRegisterRsp{
				Success: false,
				Message: "webhook is required when hook_type is 1",
			}, nil
		}
	} else {
		return &types.ServiceRegisterRsp{
			Success: false,
			Message: "invalid hook_type, must be 0 or 1",
		}, nil
	}

	// 开始事务
	err = l.db.Transaction(func(tx *gorm.DB) error {
		qTx := query.Use(tx)

		// 插入服务信息
		service := &model.ServiceInfo{
			Service:       req.Service,
			NativeAccount: req.NativeAccount,
			PdaAccount:    req.PdaAccount,
			Webhook:       req.Webhook,
			MqGroup:       req.MqGroup,
			MqTopic:       req.MqTopic,
			HookType:      req.HookType,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := qTx.ServiceInfo.WithContext(l.ctx).Create(service); err != nil {
			return err
		}

		// TODO: 插入扫描信息到t_tx_scan_info表
		// 由于t_tx_scan_info表还不存在，这里暂时不插入
		// 实际实现中应该插入扫描信息

		return nil
	})

	if err != nil {
		return &types.ServiceRegisterRsp{
			Success: false,
			Message: fmt.Sprintf("register service failed: %v", err),
		}, nil
	}

	// 启动扫描任务（异步）
	go l.startScanTaskForService(req.Service)

	return &types.ServiceRegisterRsp{
		Success: true,
		Message: "service registered successfully",
	}, nil
}

// UpdateService 更新服务
func (l *ServiceLogic) UpdateService(req *types.ServiceUpdateReq) (*types.ServiceUpdateRsp, error) {
	q := query.Use(l.db)
	serviceInfo := q.ServiceInfo

	// 检查服务是否存在
	existing, err := serviceInfo.WithContext(l.ctx).
		Where(serviceInfo.Service.Eq(req.Service)).
		First()

	if errors.Is(err, gorm.ErrRecordNotFound) || existing == nil {
		return &types.ServiceUpdateRsp{
			Success: false,
			Message: "service not found",
		}, nil
	}

	// 构建更新字段
	updateFields := make(map[string]interface{})
	updateFields["updated_at"] = time.Now()

	if req.NativeAccount != "" {
		updateFields["native_account"] = req.NativeAccount
	}
	if req.PdaAccount != "" {
		updateFields["pda_account"] = req.PdaAccount
	}
	if req.Webhook != "" {
		updateFields["webhook"] = req.Webhook
	}
	if req.MqGroup != "" {
		updateFields["mq_group"] = req.MqGroup
	}
	if req.MqTopic != "" {
		updateFields["mq_topic"] = req.MqTopic
	}
	if req.HookType >= 0 {
		updateFields["hook_type"] = req.HookType
	}

	// 验证hook_type配置
	if req.HookType == 0 {
		// 如果更新为Kafka，确保有mq_topic
		if req.MqTopic == "" && existing.MqTopic == "" {
			return &types.ServiceUpdateRsp{
				Success: false,
				Message: "mq_topic is required when hook_type is 0",
			}, nil
		}
		// Kafka不需要mq_group，可以忽略
	} else if req.HookType == 1 {
		// 如果更新为Webhook，确保有webhook
		if req.Webhook == "" && existing.Webhook == "" {
			return &types.ServiceUpdateRsp{
				Success: false,
				Message: "webhook is required when hook_type is 1",
			}, nil
		}
	}

	// 更新服务信息
	_, err = serviceInfo.WithContext(l.ctx).
		Where(serviceInfo.Service.Eq(req.Service)).
		UpdateColumns(updateFields)

	if err != nil {
		return &types.ServiceUpdateRsp{
			Success: false,
			Message: fmt.Sprintf("update service failed: %v", err),
		}, nil
	}

	return &types.ServiceUpdateRsp{
		Success: true,
		Message: "service updated successfully",
	}, nil
}

// QueryService 查询服务
func (l *ServiceLogic) QueryService(req *types.ServiceQueryReq) (*types.ServiceQueryRsp, error) {
	q := query.Use(l.db)
	serviceInfo := q.ServiceInfo

	// 查询服务信息
	service, err := serviceInfo.WithContext(l.ctx).
		Where(serviceInfo.Service.Eq(req.Service)).
		First()

	if errors.Is(err, gorm.ErrRecordNotFound) || service == nil {
		return nil, errors.New("service not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query service failed: %v", err)
	}

	// TODO: 查询扫描信息
	// 由于t_tx_scan_info表还不存在，这里暂时返回空值
	var scan *model.TxScanInfo = nil

	resp := &types.ServiceQueryRsp{
		Service:       service.Service,
		NativeAccount: service.NativeAccount,
		PdaAccount:    service.PdaAccount,
		Webhook:       service.Webhook,
		MqGroup:       service.MqGroup,
		MqTopic:       service.MqTopic,
		HookType:      service.HookType,
		CreatedAt:     service.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     service.UpdatedAt.Format(time.RFC3339),
	}

	if scan != nil {
		resp.UntilTxID = scan.UntilTxID
		resp.Slot = int64(scan.Slot)
	}

	return resp, nil
}

// startScanTaskForService 为服务启动扫描任务
func (l *ServiceLogic) startScanTaskForService(service string) {
	// 这里会启动一个goroutine来扫描该服务的交易
	// 实际实现会在TxScanTask中处理
	fmt.Printf("Starting scan task for service: %s\n", service)
}
