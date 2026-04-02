package logic

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"oshit-go/app/reward/api/types"
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
		Where(inviteRel.Invitee.Eq(nativeAccount)).
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
			inviter,
			invitee,
			channel,
			level,
			tx_id,
			created_at,
			updated_at,
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
			t.updated_at,
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
			inviter,
			invitee,
			channel,
			level,
			tx_id,
			created_at,
			updated_at,
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
			t.updated_at,
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
	if err := l.db.WithContext(l.ctx).Raw(querySQL, nativeAccount, depth).Scan(&records).Error; err != nil {
		return nil, fmt.Errorf("recursive query down invitee records error: %v", err)
	}
	return records, nil
}

func (l *RewardInviteLogic) FindInviteRelationByAccount(nativeAccount string) (*model.InviteRelation, error) {
	q := query.Use(l.db)
	inviteRel := q.InviteRelation

	record, err := inviteRel.WithContext(l.ctx).
		Where(inviteRel.Inviter.Eq(nativeAccount)).
		Or(inviteRel.Invitee.Eq(nativeAccount)).
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
	err := table.Where("inviter = ? or invitee = ?", nativeAccount, nativeAccount).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, err
}

// QueryInviterRecordByNativeAccount 通过原生地址查询邀请人信息
func (l *RewardInviteLogic) QueryInviterRecordByNativeAccount(nativeAccount string) (*model.InviteRelation, error) {
	var record model.InviteRelation
	table := l.db.Table(model.TableNameInviteRelation)
	err := table.Where("inviter = ?", nativeAccount).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, err
}

// RecordDetermineInvitationHierarchy 记录层级关系的确定信息
func (l *RewardInviteLogic) RecordDetermineInvitationHierarchy(
	inviterNativeAccount,
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
		Inviter: inviterNativeAccount,
		Invitee: inviteeNativeAccount,
		TxID:                 transferTxId,
		Channel:              inviteChannel,
		Level:                level,
	}
	err = l.db.Create(&determineInviteRecord).Error
	return &determineInviteRecord, err
}

// BuildSortedInviterItems 根据上级邀请链构建排序好的邀请人奖励列表，并裁剪到奖励层级范围内。
// 若 directInviter 不为 nil，则将其作为 index=0 的直接邀请人插入列表最前。
// 返回裁剪后的 sortedItems 和对应的 sortedClaims（与 levelRatio 对齐）。
func (l *RewardInviteLogic) BuildSortedInviterItems(
	receiptNativeAccount string,
	levelDist int32,
	levelRatio []model.LevelRatio,
	directInviter *model.NativeAccountInfo,
) ([]types.RewardTokenItem, []model.LevelRatio, error) {
	sortedInvites, err := l.GetUpInviterRecords(receiptNativeAccount, levelDist)
	if err != nil {
		return nil, nil, fmt.Errorf("recursive query up inviter records error: %w", err)
	}

	var sortedItems []types.RewardTokenItem
	for index, record := range sortedInvites {
		sortedItems = append(sortedItems, types.RewardTokenItem{
			Index:          index + 1,
			ReceiptAccount: record.Inviter,
			Amount:         0,
		})
	}

	if directInviter != nil {
		directItem := types.RewardTokenItem{
			Index:          0,
			ReceiptAccount: directInviter.NativeAccount,
			Amount:         0,
		}
		sortedItems = append([]types.RewardTokenItem{directItem}, sortedItems...)
	}

	minLen := min(len(levelRatio), len(sortedItems))
	return sortedItems[:minLen], levelRatio[:minLen], nil
}
