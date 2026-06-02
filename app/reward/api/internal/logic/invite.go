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
			inviter_level,
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
			t.inviter_level,
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
			inviter_level,
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
			t.inviter_level,
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
	// 如果To地址没有过任何邀请记录，则确定邀请层级关系
	if inviteeRecord != nil {
		return inviteeRecord, nil
	}
	// 查询邀请人作为 invitee 的记录，获取其所在层级
	// 如果邀请人本身也是被邀请进来的，则新邀请关系的 level = 邀请人的 level + 1
	// 如果邀请人不在邀请关系中（顶级用户），则 level = 1
	var inviterAsInviteeRecord model.InviteRelation
	var level int32 = 1
	if err := l.db.Table(model.TableNameInviteRelation).
		Where("invitee = ?", inviterNativeAccount).
		First(&inviterAsInviteeRecord).Error; err == nil {
		level = inviterAsInviteeRecord.InviterLevel + 1
	}
	determineInviteRecord := model.InviteRelation{
		Inviter:      inviterNativeAccount,
		Invitee:      inviteeNativeAccount,
		TxID:         transferTxId,
		Channel:      inviteChannel,
		InviterLevel: level,
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
	var sortedItems []types.RewardTokenItem

	if directInviter != nil {
		// 首次建立邀请关系时，receiptAccount 尚未写入 t_invite_relation（Kafka 确认后才写入），
		// 从 receiptAccount 开始的 CTE 查不到任何记录。
		// 因此改为从 directInviter 开始向上查邀请链，再将 directInviter 自身前插到列表首位。
		upInvites, err := l.GetUpInviterRecords(directInviter.NativeAccount, levelDist)
		if err != nil {
			return nil, nil, fmt.Errorf("recursive query up inviter records error: %w", err)
		}
		sortedItems = append(sortedItems, types.RewardTokenItem{
			Index:          0,
			ReceiptAccount: directInviter.NativeAccount,
			Amount:         0,
		})
		for index, record := range upInvites {
			sortedItems = append(sortedItems, types.RewardTokenItem{
				Index:          index + 1,
				ReceiptAccount: record.Inviter,
				Amount:         0,
			})
		}
	} else {
		// 邀请关系已存在，直接从 receiptAccount 向上查完整邀请链
		sortedInvites, err := l.GetUpInviterRecords(receiptNativeAccount, levelDist)
		if err != nil {
			return nil, nil, fmt.Errorf("recursive query up inviter records error: %w", err)
		}
		for index, record := range sortedInvites {
			sortedItems = append(sortedItems, types.RewardTokenItem{
				Index:          index + 1,
				ReceiptAccount: record.Inviter,
				Amount:         0,
			})
		}
	}

	minLen := min(len(levelRatio), len(sortedItems))
	return sortedItems[:minLen], levelRatio[:minLen], nil
}
