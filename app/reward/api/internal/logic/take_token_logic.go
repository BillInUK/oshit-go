package logic

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	rewardrpc "oshit-go/app/reward/api/internal/rpc"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/app/reward/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/constants"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
	"runtime/debug"
	"time"
)

type TakeTokenLogic struct {
	prefix            string
	decimals          uint8
	ctx               context.Context
	srvCtx            *svc.ServiceContext
	db                *gorm.DB
	rd                redis.UniversalClient
	rs                redsync.Redsync
	baseClient        *rewardrpc.BaseClient
	levelDist         int32
	levelRatio        []model.LevelRatio
	levelRatioMap     map[int32]model.LevelRatio
	serviceConfig     *model.TakeTokenConfig
	rpcClient         *rpc.Client
	LightHouseAddress solana.PublicKey
	inviteLogic       *RewardInviteLogic
}

func NewTakeLogic(ctx context.Context, srvCtx *svc.ServiceContext) *TakeTokenLogic {
	return &TakeTokenLogic{
		prefix:            "TakeToken -",
		ctx:               ctx,
		srvCtx:            srvCtx,
		db:                srvCtx.DB,
		rd:                srvCtx.Redis,
		rs:                srvCtx.RedSync,
		rpcClient:         srvCtx.RpcClient,
		baseClient:        srvCtx.BaseClient,
		levelDist:         srvCtx.LevelDist.Level,
		levelRatio:        srvCtx.LevelRatio,
		levelRatioMap:     srvCtx.LevelRatioMap,
		serviceConfig:     srvCtx.TakeTokenConfig,
		decimals:          uint8(srvCtx.TokenConfig.Decimals),
		LightHouseAddress: srvCtx.LightHouseAddress,
		inviteLogic:       NewRewardInviteLogic(ctx, srvCtx.DB),
	}
}

// GetConfig 获取奖励规则配置
func (l *TakeTokenLogic) GetConfig() (*model.TakeTokenConfig, error) {
	return l.serviceConfig, nil
}

// GetRecordByTxId 根据交易id获取领取记录
func (l *TakeTokenLogic) GetRecordByTxId(txId string) (*model.TakeTokenRecord, error) {
	var record model.TakeTokenRecord
	db := l.db.Model(record)
	if err := db.Where("tx_id = ?", txId).First(&record).Error; err != nil {
		return nil, err
	}
	return nil, nil
}

// inviteCodeValid 判断邀请码是否有效
// 必须确保 邀请码有效获取确定邀请关系时会返回直接邀请人信息
func (l *TakeTokenLogic) inviteCodeValid(ctx context.Context, receiptAccount, inviteCode string) (*model.NativeAccountInfo, bool, bool, error) {
	var prefix = fmt.Sprintf("%s 根据地址 [%s]  邀请码 [%s] 获取交易信息 -", l.prefix, receiptAccount, inviteCode)
	// 1. 邀请码为空，直接跳过查询
	if inviteCode == "" {
		log.Infof("%s 邀请码无效 - 邀请码为空，跳过邀请人查询", prefix)
		return nil, false, false, nil
	}
	// 2. 直接邀请人不存在，或使用自己的邀请码
	directInviter, err := l.inviteLogic.GetAccountByInviteCode(inviteCode)
	if err != nil {
		log.Errorf("%s 邀请码无效 - 查询邀请人失败: %v", prefix, err)
		return nil, false, false, err
	}
	if directInviter == nil || directInviter.NativeAccount == receiptAccount {
		log.Infof("%s 邀请码无效 - 邀请人不存在或者使用自己的邀请码来领取奖励", prefix)
		return nil, false, false, nil
	}
	// 3. 查看是否有过通过邀请码领取成功的记录，如果有过邀请记录，则邀请码无效，不确定邀请关系
	takeRecord, err := l.GetRecordByInviteCode(receiptAccount)
	if err != nil {
		log.Errorf("%s 查询是否有邀请记录错误: %v", prefix, err)
		return directInviter, false, false, errors.New("get token token record error")
	}
	// 如果有过邀请码领取记录，则邀请码无效
	if takeRecord != nil {
		log.Infof("%s 邀请码无效 - 有过使用邀请码领取记录", prefix)
		return directInviter, false, false, nil
	}
	// 4. 查询地址是否存在于邀请关系表里面，如果存在，则不确定邀请关系
	receiptInviteRecord, err := l.inviteLogic.FindInviteRelationByAccount(receiptAccount)
	if err != nil {
		log.Errorf("%s 查询地址是否在邀请关系内错误 :%v", prefix, err)
		return directInviter, false, false, errors.New("query invite record exist by to native account error")
	}
	if receiptInviteRecord != nil && receiptInviteRecord.RecordID != "" {
		log.Infof("%s 邀请码无效 - 地址已经邀请关系内", prefix)
		return directInviter, true, false, nil
	}
	return directInviter, true, true, nil
}

// getTxInfo 获取take token的交易信息
func (l *TakeTokenLogic) getTxInfo(ctx context.Context, receiptNativeAccount, inviteCode string) (*types.TakeTokenTxInfo, error) {
	var prefix = fmt.Sprintf("%s 根据地址 [%s]  邀请码 [%s] 获取交易信息 -", l.prefix, receiptNativeAccount, inviteCode)
	var takeTxInfo types.TakeTokenTxInfo

	// 1. 确定邀请码是否有效，是否建立邀请关系
	directInviter, codeValid, invited, err := l.inviteCodeValid(ctx, receiptNativeAccount, inviteCode)
	if err != nil {
		log.Errorf("%s 确定是否建立邀请关系错误: %v", prefix, err)
		return nil, err
	}

	// 2. 根据邀请码是否有效决定领取奖励的金额
	var rewardAmount uint64 = 0
	if !codeValid {
		rewardAmount = uint64(l.serviceConfig.Amount)
	} else {
		rewardAmount = uint64(l.serviceConfig.InviteAmount)
	}

	// 3. 获取领取地址的nativeAccount的tokenAccount
	receiptTokenPubKey, _, _ := solana.FindAssociatedTokenAddress(solana.MPK(receiptNativeAccount), solana.MPK(l.serviceConfig.TokenMintAccount))
	receiptTokenAccount := receiptTokenPubKey.String()

	// 如果确定邀请关系
	var sortedInvites []model.InviteRelation
	var sortedItems []types.RewardTokenItem
	if (invited || codeValid) && directInviter != nil {
		// 如果确定邀请·，则查询邀请码对应的地址的的上级和上上级
		sortedInvites, err = l.inviteLogic.GetUpInviterRecords(receiptNativeAccount, l.levelDist)
		if err != nil {
			log.Errorf("%s 递归向上查询邀请人错误[%v]", prefix, err)
			return nil, errors.New("recursive query up inviter records error")
		}
		// 排序排序邀请人信息，加上索引，方便前端排序
		for index, record := range sortedInvites {
			account := types.RewardTokenItem{
				Index:         index + 1,
				NativeAccount: record.InviterNativeAccount,
				TokenAccount:  record.InviterTokenAccount,
				Amount:        uint64(0),
			}
			sortedItems = append(sortedItems, account)
		}
		// 将邀请人放到邀请账户记录之前
		directItem := types.RewardTokenItem{Index: 0, NativeAccount: directInviter.NativeAccount, TokenAccount: directInviter.TokenAccount, Amount: uint64(0)}
		sortedItems = append([]types.RewardTokenItem{directItem}, sortedItems...)
	} else {
		// 如果不确定邀请关系，则查询领取奖励地址的上级和上上级
		sortedInvites, err = l.inviteLogic.GetUpInviterRecords(receiptNativeAccount, l.levelDist)
		if err != nil {
			log.Errorf("%s 递归向上查询邀请人错误[%v]", prefix, err)
			return nil, errors.New("recursive query up inviter records error")
		}
		for index, record := range sortedInvites {
			account := types.RewardTokenItem{
				Index:         index + 1,
				NativeAccount: record.InviterNativeAccount,
				TokenAccount:  record.InviterTokenAccount,
				Amount:        uint64(0),
			}
			sortedItems = append(sortedItems, account)
		}
	}

	// 获取到要奖励的邀请人之后，取奖励层级和邀请人的最小集合
	minLen := min(len(l.levelRatio), len(sortedItems))
	sortedClaims := l.levelRatio[:minLen]
	sortedItems = sortedItems[:minLen]

	// 填写领取奖励信息
	rewardInfo := types.RewardTokenItem{Index: 0, NativeAccount: receiptNativeAccount, TokenAccount: receiptTokenAccount, Amount: rewardAmount}
	// 在最小集合里面决定每个层级的邀请人领取多少金额
	for index, claim := range sortedClaims {
		sortedItems[index].Amount = uint64(float64(rewardAmount) * claim.Ratio)
	}
	// 返回领取奖励信息
	takeTxInfo.RewardNativeAccount = l.serviceConfig.RewardNativeAccount
	takeTxInfo.RewardTokenAccount = l.serviceConfig.RewardTokenAccount
	takeTxInfo.TokenMintAccount = l.serviceConfig.TokenMintAccount
	takeTxInfo.DexNativeAccount = l.serviceConfig.DexNativeAccount
	takeTxInfo.DexFeeRate = l.serviceConfig.DexFeeRate
	takeTxInfo.MaxDexFee = l.serviceConfig.MaxDexFee
	takeTxInfo.Decimals = l.serviceConfig.Decimals
	takeTxInfo.InviteCode = inviteCode
	takeTxInfo.RewardInfo = rewardInfo
	takeTxInfo.InviteCodeValid = codeValid

	// 获取本次奖励的总代币量
	totalRewardAmount := takeTxInfo.RewardInfo.Amount
	for _, item := range takeTxInfo.RewardInviterInfo {
		totalRewardAmount += item.Amount
	}
	// 计算所有的token的价格
	quoteSOLPrice, err := l.baseClient.GetTokenQuoteSOLPrice(ctx)
	if err != nil {
		log.Errorf("%s 计算奖励金额价格错误: %v", prefix, err)
		return nil, fmt.Errorf("get toke quote sol price failed")
	}
	takeTxInfo.QuoteSOLPrice = quoteSOLPrice
	takeTxInfo.TotalRewardAmount = float64(totalRewardAmount)
	takeTxInfo.QuotedSOLAmount = quoteSOLPrice * float64(totalRewardAmount) / l.srvCtx.TokenDecimal * float64(solana.LAMPORTS_PER_SOL)
	takeTxInfo.InviteDetermine = invited
	takeTxInfo.Claims = sortedClaims
	takeTxInfo.RewardInviterInfo = sortedItems

	return &takeTxInfo, nil
}

// ProcessGetTxInfo 获取take token的交易信息
func (l *TakeTokenLogic) ProcessGetTxInfo(ctx context.Context, req types.GetTakeTokenTxInfoReq) (*types.TakeTokenTxInfo, error) {
	return l.getTxInfo(ctx, req.ReceiptAccount, req.InviteCode)
}

// GetRecordByInviteCode 查看地址是否有有过使用邀请码take token的纪录
func (l *TakeTokenLogic) GetRecordByInviteCode(nativeAccount string) (*model.TakeTokenRecord, error) {
	var err error
	var record model.TakeTokenRecord
	table := l.db.Table(model.TableNameTakeTokenRecord)
	err = table.Where("receipt_native_account = ? and use_invite_code = ? and state = ?", nativeAccount, true, 1).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, nil
}

// sendTransaction 签名并广播用户提交上来的交易
func (l *TakeTokenLogic) sendTransaction(ctx context.Context, tx *solana.Transaction, service string) (*solana.Signature, error) {
	messageContent, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("encode transaction message for signing error:%v", err)
	}
	privateKey, exist := l.srvCtx.RewardKeyMap[service]
	if !exist {
		return nil, fmt.Errorf("sign tx error can not find %s private key", service)
	}
	rewardSign, err := privateKey.Sign(messageContent)
	if err != nil {
		return nil, fmt.Errorf("failed to signed with reward error:%v", err)
	}
	if len(tx.Signatures) == 2 {
		tx.Signatures[1] = rewardSign
	}
	if len(tx.Signatures) == 1 {
		tx.Signatures = append(tx.Signatures, rewardSign)
	}
	// 广播交易
	txId, err := l.rpcClient.SendTransaction(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("send transaction error:%v", err)
	}
	return &txId, nil
}

// RecordOfficialGiveTokenRecord 记录官网领取token记录
func (l *TakeTokenLogic) recordTakeToken(takeTokenTxInfo *types.TakeTokenTxInfo, decodedServiceTx *entity.DecodedServiceTransaction, invited bool) error {
	// 开启事务
	dbTx := l.db.Begin()
	if dbTx.Error != nil {
		log.Errorf("官网领取奖励 - 开启事务错误: %v", dbTx.Error)
		return dbTx.Error
	}

	// 记录领取的交易ID
	txRecord := model.ServiceTx{
		Service:    "Reward",
		SubService: "TakeToken",
		TxID:       decodedServiceTx.TxID,
		CreatedAt:  time.Now(),
	}
	if err := dbTx.Table(model.TableNameServiceTx).Create(&txRecord).Error; err != nil {
		dbTx.Rollback() // 回滚事务
		log.Errorf("官网领取奖励 - 插入交易业务类型表错误: %v", err)
		return err
	}

	// 记录领取奖励记录
	takeTokenRecord := model.TakeTokenRecord{
		TokenMintAccount:     decodedServiceTx.RewardInst.TokenMintAccount,
		RewardTokenAccount:   decodedServiceTx.RewardInst.FromTokenAccount,
		RewardNativeAccount:  decodedServiceTx.RewardInst.FromNativeAccount,
		ReceiptTokenAccount:  decodedServiceTx.RewardInst.ToTokenAccount,
		ReceiptNativeAccount: decodedServiceTx.RewardInst.ToNativeAccount,
		DexNativeAccount:     decodedServiceTx.ToDexInst.ToNativeAccount,
		RewardTxID:           decodedServiceTx.TxID,
		Amount:               l.serviceConfig.Amount,
		DexFee:               decodedServiceTx.ToDexInst.Amount,
		UseInviteCode:        takeTokenTxInfo.InviteCodeValid,
		InviteCode:           takeTokenTxInfo.InviteCode,
		State:                0,
		Invited:              invited,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	if err := dbTx.Table(model.TableNameTakeTokenRecord).Create(&takeTokenRecord).Error; err != nil {
		dbTx.Rollback() // 回滚事务
		log.Errorf("官网领取奖励 - 插入领取记录表错误: %v", err)
		return err
	}
	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("官网领取奖励 - 提交事务错误: %v", err)
		return err
	}

	return nil
}

// recordFundFlow 记录官网领取token记录
func (l *TakeTokenLogic) recordFundFlow(brand, tokenSymbol string, decodedServiceTx *entity.DecodedServiceTransaction) error {
	var err error
	var fundFlows []model.SolFundFlow

	table := l.db.Table(model.TableNameSolFundFlow)

	// 1.记录dex入账sol流水
	log.Infof("记录官网领取token流水 - 记录dex入账sol流水 from %v to %v", decodedServiceTx.ToDexInst.FromNativeAccount, decodedServiceTx.ToDexInst.ToNativeAccount)

	dexInputFlow := model.SolFundFlow{
		Brand:             brand,
		TokenSymbol:       tokenSymbol,
		IsToken:           false,
		FromNativeAccount: decodedServiceTx.FromNativeAccount,
		ToNativeAccount:   decodedServiceTx.ToDexInst.ToNativeAccount,
		TxID:              decodedServiceTx.TxID,
		Direction:         constants.FlowInput,
		ServiceType:       constants.ServiceOfficialGive,
		FlowType:          constants.FlowOfficialGiveDexFee,
		Decimals:          9,
		Amount:            float64(decodedServiceTx.ToDexInst.Amount),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	fundFlows = append(fundFlows, dexInputFlow)

	// 2.记录奖励转账人出账流水
	log.Infof("记录官网领取token流水 - 记录奖励转账人出账流水 from %v to %v amount %v",
		decodedServiceTx.RewardInst.FromNativeAccount, decodedServiceTx.RewardInst.ToNativeAccount,
		decodedServiceTx.RewardInst.Amount)

	rewardOutputFlow := model.SolFundFlow{
		Brand:             brand,
		TokenSymbol:       tokenSymbol,
		IsToken:           true,
		FromNativeAccount: decodedServiceTx.RewardInst.FromNativeAccount,
		ToNativeAccount:   decodedServiceTx.RewardInst.ToNativeAccount,
		TxID:              decodedServiceTx.TxID,
		Direction:         constants.FlowOutput,
		ServiceType:       constants.ServiceOfficialGive,
		FlowType:          constants.FlowOfficialGiveTokenReward,
		Decimals:          int16(decodedServiceTx.RewardInst.Decimals),
		Amount:            decodedServiceTx.RewardInst.Amount,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	fundFlows = append(fundFlows, rewardOutputFlow)

	// 3.记录奖励转账人上级邀请人出账流水
	for _, inst := range decodedServiceTx.RewardInviterInst {
		log.Infof("记录官网领取token流水 - 记录奖励转账人上级邀请人出账流水 from %v to %v amount %v",
			inst.FromNativeAccount, inst.ToNativeAccount, inst.Amount)

		outputFlow := model.SolFundFlow{
			Brand:             brand,
			TokenSymbol:       tokenSymbol,
			IsToken:           true,
			FromNativeAccount: inst.FromNativeAccount,
			ToNativeAccount:   inst.ToNativeAccount,
			TxID:              decodedServiceTx.TxID,
			Direction:         constants.FlowOutput,
			ServiceType:       constants.ServiceOfficialGive,
			FlowType:          constants.FlowOfficialGiveTokenRewardInviter,
			Decimals:          int16(inst.Decimals),
			Amount:            inst.Amount,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		fundFlows = append(fundFlows, outputFlow)
	}

	// 批量插入，当唯一键冲突时更新UpdateTime
	batchSize := 100

	// 使用ON CONFLICT DO UPDATE SET (推荐)
	if err = table.Omit("record_id").Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tx_id"},
			{Name: "to_native_account"},
			{Name: "flow_type"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"updated_at": time.Now(),
		}),
	}).CreateInBatches(&fundFlows, batchSize).Error; err != nil {
		log.Errorf("TakeToken - 记录流水错误: %v", err)
		return err
	}

	return nil
}

// ProcessCommitTx 处理提交上来的交易
func (l *TakeTokenLogic) ProcessCommitTx(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx, inviteCode string) (*solana.Signature, error) {
	takeTxInfo, err := l.getTxInfo(ctx, preCheckedTx.From.String(), inviteCode)
	if err != nil {
		return nil, fmt.Errorf("get take token transaction info error: %v", err)
	}

	// 解析出来交易里面的transfer checked和transfer指令集合
	decodedTx, err := l.decodeSOLTx(takeTxInfo, &preCheckedTx.SOLTx)
	if err != nil {
		log.Errorf("%s 解析solana交易错误: %v", l.prefix, err)
		return nil, fmt.Errorf("decode transaction error:%s", err)
	}

	// 检查 transfer checked 指令和 transfer 指令是否符合奖励要求
	decodedServiceTx, err := l.checkDecodedSOLTx(takeTxInfo, decodedTx)
	if err != nil {
		log.Errorf("%s 校验solana交易当中的指令错误: %v", l.prefix, err)
		return nil, fmt.Errorf("check transaction instruction failed: %v", err)
	}

	// 通过 base 模块签名并异步广播
	txIdStr, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Reward", "TakeToken")
	if err != nil {
		log.Errorf("%s 发送交易失败,错误: %v", l.prefix, err)
		return nil, errors.New(utils.FilterAndTranslateSOLError(err))
	}
	txId := solana.MustSignatureFromBase58(txIdStr)

	// 记录领取记录到数据库
	decodedServiceTx.TxID = txIdStr
	if err = l.recordTakeToken(takeTxInfo, decodedServiceTx, takeTxInfo.InviteDetermine); err != nil {
		log.Errorf("%s 记录交易信息,错误: %v", l.prefix, err)
		return &txId, errors.New("record official transfer token error")
	}

	return &txId, nil
}

// HandleScannedTx 处理 base 模块推送的 TakeToken 链上已确认交易
func (l *TakeTokenLogic) HandleScannedTx(msg entity.NewScannedTx) error {
	txId := msg.TxSig.Signature.String()
	prefix := fmt.Sprintf("%s 处理扫描到的交易 %s -", l.prefix, txId)

	// 开始事务
	dbTx := l.db.Begin()

	// 确保在函数结束时进行事务的提交或回滚
	defer func() {
		if r := recover(); r != nil {
			// 打印堆栈信息
			stack := debug.Stack()
			// 打印错误日志和堆栈信息
			log.Errorf("%s 处理消息时发生错误: %v\n堆栈信息:\n%s", prefix, r, string(stack))
			// 回滚事务
			dbTx.Rollback()
		}
	}()
	var err error
	var takeTokenRecord model.TakeTokenRecord

	// 根据交易 Id 找到记录
	table := dbTx.Table(model.TableNameTakeTokenRecord)
	if err = table.Where("reward_tx_id = ?", txId).First(&takeTokenRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("%s 查找业务记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}

	if msg.TxSig.Err != nil {
		if err = dbTx.Table(model.TableNameTakeTokenRecord).
			Where("record_id = ?", takeTokenRecord.RecordID).Update("state", -1).Error; err != nil {
			log.Errorf("%s 更新领取交易状态为失败，错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
	} else {
		if err = dbTx.Table(model.TableNameTakeTokenRecord).
			Where("record_id = ?", takeTokenRecord.RecordID).Update("state", 1).Error; err != nil {
			log.Errorf("%s 更新领取交易状态为成功，错误: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
		// 记录流水
		decodedServiceTx, err := app_utils.DecodeServiceTransaction(&msg.DecodedTx)
		if err != nil {
			log.Errorf("%s 解码服务交易失败: %v", prefix, err)
			dbTx.Rollback()
			return err
		}
		err = l.recordFundFlow("OShit", "OShit", decodedServiceTx)
		if err != nil {
			log.Errorf("%s 记录流水失败: %v", prefix, err)
			dbTx.Rollback()
			return err
		}

		log.Infof("%s 更新记录为成功，邀请码 %v，确定邀请关系 %v", prefix, takeTokenRecord.InviteCode, takeTokenRecord.Invited)

		// 如果需要确定邀请层级关系
		if takeTokenRecord.Invited {
			inviterInfo, err := l.inviteLogic.GetAccountByInviteCode(takeTokenRecord.InviteCode)
			if err != nil {
				log.Errorf("%s 确定邀请关系错误，无法根据邀请码找到邀请人", prefix)
				dbTx.Rollback()
				return err
			}
			if inviterInfo != nil {
				_, err = l.inviteLogic.RecordDetermineInvitationHierarchy(
					inviterInfo.TokenAccount,
					inviterInfo.NativeAccount,
					takeTokenRecord.ReceiptTokenAccount,
					takeTokenRecord.ReceiptNativeAccount,
					takeTokenRecord.RewardTxID,
					"InviteCode",
				)
				if err != nil {
					log.Errorf("%s 记录邀请层级关系错误: %v", prefix, err)
					dbTx.Rollback()
					return err
				}
			}
		}
	}

	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("官网领取奖励 - 提交事务失败: %v", err)
		dbTx.Rollback()
		return err
	}
	return nil
}

// HandleExpiredTx 处理 base 模块推送的 TakeToken 已超时交易
func (l *TakeTokenLogic) HandleExpiredTx(msg entity.NewExpiredTx) error {
	txId := msg.TxID
	prefix := fmt.Sprintf("%s 处理超时交易 %s -", l.prefix, txId)

	// 开始事务
	dbTx := l.db.Begin()

	// 确保在函数结束时进行事务的提交或回滚
	defer func() {
		if r := recover(); r != nil {
			// 打印堆栈信息
			stack := debug.Stack()
			// 打印错误日志和堆栈信息
			log.Errorf("%s 处理消息时发生错误: %v\n堆栈信息:\n%s", prefix, r, string(stack))
			// 回滚事务
			dbTx.Rollback()
		}
	}()
	var err error
	var takeTokenRecord model.TakeTokenRecord

	// 根据交易 Id 找到记录
	table := dbTx.Table(model.TableNameTakeTokenRecord)
	if err = table.Where("reward_tx_id = ?", txId).First(&takeTokenRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dbTx.Rollback()
			return err
		}
		log.Errorf("%s 查找领取记录错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}
	if err = table.Where("reward_tx_id = ?", txId).Update("state", -1).Error; err != nil {
		log.Errorf("%s 更新领取交易状态为失败，错误: %v", prefix, err)
		dbTx.Rollback()
		return err
	}
	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("%s 提交事务失败: %v", prefix, err)
		dbTx.Rollback()
		return err
	}
	return nil
}
