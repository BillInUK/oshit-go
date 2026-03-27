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

type InviteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	db     *gorm.DB
}

func NewInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InviteLogic {
	return &InviteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		db:     svcCtx.DB,
	}
}

func (l *InviteLogic) GetAccountByInviteCode(req *types.GetAccountByInviteCodeReq) (*types.GetAccountByInviteCodeRsp, error) {
	q := query.Use(l.svcCtx.DB)
	naInfo := q.NativeAccountInfo

	record, err := naInfo.WithContext(l.ctx).
		Where(naInfo.InviteCode.Eq(req.InviteCode)).
		First()

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("invite code not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query native account info error: %v", err)
	}

	return &types.GetAccountByInviteCodeRsp{
		RecordID:      record.RecordID,
		NativeAccount: record.NativeAccount,
		TokenAccount:  record.TokenAccount,
		InviteCode:    record.InviteCode,
		CreatedAt:     record.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (l *InviteLogic) CheckInviteRecord(req *types.CheckInviteRecordReq) (*types.CheckInviteRecordRsp, error) {
	q := query.Use(l.svcCtx.DB)
	inviteRel := q.InviteRelation

	count, err := inviteRel.WithContext(l.ctx).
		Where(inviteRel.Invitee.Eq(req.NativeAccount)).
		Count()

	if err != nil {
		return nil, fmt.Errorf("check invite record error: %v", err)
	}

	return &types.CheckInviteRecordRsp{
		Exists: count > 0,
	}, nil
}

func (l *InviteLogic) GetUpInviterRecords(req *types.RecursiveQueryReq) (*types.RecursiveQueryRsp, error) {
	if req.Depth <= 0 || req.Depth > 20 {
		return nil, errors.New("depth must be between 1 and 20")
	}

	querySQL := `
	WITH RECURSIVE invite_tree AS (
		SELECT
			record_id,
			inviter,
			invitee,
			channel,
			level,
			tx_id,
			created_at,
			1 AS depth
		FROM
			public.t_invite_relation
		WHERE
			invitee = ?

		UNION ALL

		SELECT
			t.record_id,
			t.inviter,
			t.invitee,
			t.channel,
			t.level,
			t.tx_id,
			t.created_at,
			it.depth + 1 AS depth
		FROM
			public.t_invite_relation t
		INNER JOIN
			invite_tree it ON t.invitee = it.inviter
		WHERE
			it.depth < ?
	)
	SELECT * FROM invite_tree;
	`

	var records []model.InviteRelation
	err := l.db.WithContext(l.ctx).Raw(querySQL, req.NativeAccount, req.Depth).Scan(&records).Error
	if err != nil {
		return nil, fmt.Errorf("recursive query up inviter records error: %v", err)
	}

	responseRecords := make([]types.InviteRelation, 0, len(records))
	for _, record := range records {
		responseRecords = append(responseRecords, types.InviteRelation{
			RecordID:  record.RecordID,
			Inviter:   record.Inviter,
			Invitee:   record.Invitee,
			Channel:   record.Channel,
			Level:     record.Level,
			TxID:      record.TxID,
			CreatedAt: record.CreatedAt.Format(time.RFC3339),
		})
	}

	return &types.RecursiveQueryRsp{
		Records: responseRecords,
	}, nil
}

func (l *InviteLogic) GetDownInviteeRecords(req *types.RecursiveQueryReq) (*types.RecursiveQueryRsp, error) {
	if req.Depth <= 0 || req.Depth > 20 {
		return nil, errors.New("depth must be between 1 and 20")
	}

	querySQL := `
	WITH RECURSIVE invite_tree AS (
		SELECT
			record_id,
			inviter,
			invitee,
			channel,
			level,
			tx_id,
			created_at,
			1 AS depth
		FROM
			public.t_invite_relation
		WHERE
			inviter = ?

		UNION ALL

		SELECT
			t.record_id,
			t.inviter,
			t.invitee,
			t.channel,
			t.level,
			t.tx_id,
			t.created_at,
			it.depth + 1 AS depth
		FROM
			public.t_invite_relation t
		INNER JOIN
			invite_tree it ON t.inviter = it.invitee
		WHERE
			it.depth < ?
	)
	SELECT * FROM invite_tree;
	`

	var records []model.InviteRelation
	err := l.db.WithContext(l.ctx).Raw(querySQL, req.NativeAccount, req.Depth).Scan(&records).Error
	if err != nil {
		return nil, fmt.Errorf("recursive query down invitee records error: %v", err)
	}

	responseRecords := make([]types.InviteRelation, 0, len(records))
	for _, record := range records {
		responseRecords = append(responseRecords, types.InviteRelation{
			RecordID:  record.RecordID,
			Inviter:   record.Inviter,
			Invitee:   record.Invitee,
			Channel:   record.Channel,
			Level:     record.Level,
			TxID:      record.TxID,
			CreatedAt: record.CreatedAt.Format(time.RFC3339),
		})
	}

	return &types.RecursiveQueryRsp{
		Records: responseRecords,
	}, nil
}
