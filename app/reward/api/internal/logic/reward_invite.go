package logic

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/dal/query"
)

type RewardInviteLogic struct {
	ctx context.Context
	db  *gorm.DB
}

func NewRewardInviteLogic(ctx context.Context, db *gorm.DB) *RewardInviteLogic {
	return &RewardInviteLogic{ctx: ctx, db: db}
}

func (l *RewardInviteLogic) GetAccountByInviteCode(inviteCode string) (*model.NativeAccountInfo, error) {
	q := query.Use(l.db)
	naInfo := q.NativeAccountInfo

	record, err := naInfo.WithContext(l.ctx).
		Where(naInfo.InviteCode.Eq(inviteCode)).
		First()

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query native account info error: %v", err)
	}
	return record, nil
}

func (l *RewardInviteLogic) CheckInviteRecord(nativeAccount string) (bool, error) {
	q := query.Use(l.db)
	inviteRel := q.InviteRelation

	count, err := inviteRel.WithContext(l.ctx).
		Where(inviteRel.InviteeNativeAccount.Eq(nativeAccount)).
		Count()

	if err != nil {
		return false, fmt.Errorf("check invite record error: %v", err)
	}
	return count > 0, nil
}

func (l *RewardInviteLogic) GetUpInviterRecords(nativeAccount string, depth int32) ([]model.InviteRelation, error) {
	if depth <= 0 || depth > 20 {
		return nil, errors.New("depth must be between 1 and 20")
	}

	querySQL := `
	WITH RECURSIVE invite_tree AS (
		SELECT
			record_id,
			inviter_native_account,
			inviter_token_account,
			invitee_native_account,
			invitee_token_account,
			channel,
			level,
			tx_id,
			created_at,
			updated_at,
			1 AS depth
		FROM
			public.t_invite_relation
		WHERE
			invitee_native_account = ?

		UNION ALL

		SELECT
			t.record_id,
			t.inviter_native_account,
			t.inviter_token_account,
			t.invitee_native_account,
			t.invitee_token_account,
			t.channel,
			t.level,
			t.tx_id,
			t.created_at,
			t.updated_at,
			it.depth + 1 AS depth
		FROM
			public.t_invite_relation t
		INNER JOIN
			invite_tree it ON t.invitee_native_account = it.inviter_native_account
		WHERE
			it.depth < ?
	)
	SELECT * FROM invite_tree;
	`

	var records []model.InviteRelation
	if err := l.db.WithContext(l.ctx).Raw(querySQL, nativeAccount, depth).Scan(&records).Error; err != nil {
		return nil, fmt.Errorf("recursive query up inviter records error: %v", err)
	}
	return records, nil
}

func (l *RewardInviteLogic) GetDownInviteeRecords(nativeAccount string, depth int32) ([]model.InviteRelation, error) {
	if depth <= 0 || depth > 20 {
		return nil, errors.New("depth must be between 1 and 20")
	}

	querySQL := `
	WITH RECURSIVE invite_tree AS (
		SELECT
			record_id,
			inviter_native_account,
			inviter_token_account,
			invitee_native_account,
			invitee_token_account,
			channel,
			level,
			tx_id,
			created_at,
			updated_at,
			1 AS depth
		FROM
			public.t_invite_relation
		WHERE
			inviter_native_account = ?

		UNION ALL

		SELECT
			t.record_id,
			t.inviter_native_account,
			t.inviter_token_account,
			t.invitee_native_account,
			t.invitee_token_account,
			t.channel,
			t.level,
			t.tx_id,
			t.created_at,
			t.updated_at,
			it.depth + 1 AS depth
		FROM
			public.t_invite_relation t
		INNER JOIN
			invite_tree it ON t.inviter_native_account = it.invitee_native_account
		WHERE
			it.depth < ?
	)
	SELECT * FROM invite_tree;
	`

	var records []model.InviteRelation
	if err := l.db.WithContext(l.ctx).Raw(querySQL, nativeAccount, depth).Scan(&records).Error; err != nil {
		return nil, fmt.Errorf("recursive query down invitee records error: %v", err)
	}
	return records, nil
}

func (l *RewardInviteLogic) FindInviteRelationByAccount(nativeAccount string) (*model.InviteRelation, error) {
	q := query.Use(l.db)
	inviteRel := q.InviteRelation

	record, err := inviteRel.WithContext(l.ctx).
		Where(inviteRel.InviterNativeAccount.Eq(nativeAccount)).
		Or(inviteRel.InviteeNativeAccount.Eq(nativeAccount)).
		First()

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find invite relation by account error: %v", err)
	}
	return record, nil
}

// InviteRelationExist 通过原生地址查找是否地址是否在层级关系表里面是否已经存在记录
func (l *RewardInviteLogic) InviteRelationExist(nativeAccount string) (*model.InviteRelation, error) {
	var record model.InviteRelation
	table := l.db.Table(model.TableNameInviteRelation)
	err := table.Where("inviter_native_account = ? or invitee_native_account = ?", nativeAccount, nativeAccount).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, err
}

// QueryInviterRecordByNativeAccount 通过原生地址查询邀请人信息
func (l *RewardInviteLogic) QueryInviterRecordByNativeAccount(nativeAccount string) (*model.InviteRelation, error) {
	var record model.InviteRelation
	table := l.db.Table(model.TableNameInviteRelation)
	err := table.Where("inviter_native_account = ?", nativeAccount).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, err
}

// RecordDetermineInvitationHierarchy 记录层级关系的确定信息
func (l *RewardInviteLogic) RecordDetermineInvitationHierarchy(
	inviterTokenAccount,
	inviterNativeAccount,
	inviteeTokenAccount,
	inviteeNativeAccount,
	transferTxId,
	inviteChannel string) (*model.InviteRelation, error) {

	if transferTxId == "" && inviteChannel != "InviteLink" {
		return nil, errors.New("transfer tx id is null and invite channel is not InviteLink")
	}
	// 查看to地址是否在邀请层级关系里面
	inviteeRecord, err := l.InviteRelationExist(inviteeNativeAccount)
	if err != nil {
		return nil, err
	}
	inviterRecord, err := l.QueryInviterRecordByNativeAccount(inviterNativeAccount)
	if err != nil {
		return nil, err
	}
	// 如果To地址没有过任何邀请记录，则确定邀请层级关系
	if inviteeRecord != nil {
		return inviteeRecord, nil
	}
	// 如果邀请人没有在邀请关系里面，则Level为1，如果邀请人已经在邀请关系里面了，则Level为邀请人的层级+1
	var level int32 = 1
	if inviterRecord != nil {
		level = inviterRecord.Level + 1
	}
	determineInviteRecord := model.InviteRelation{
		InviterTokenAccount:  inviterTokenAccount,
		InviterNativeAccount: inviterNativeAccount,
		InviteeTokenAccount:  inviteeTokenAccount,
		InviteeNativeAccount: inviteeNativeAccount,
		TxID:                 transferTxId,
		Channel:              inviteChannel,
		Level:                level,
	}
	err = l.db.Create(&determineInviteRecord).Error
	return &determineInviteRecord, err
}
