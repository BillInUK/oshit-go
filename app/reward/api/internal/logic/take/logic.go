package take

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
	"oshit-go/app/reward/api/internal/logic"
	rewardrpc "oshit-go/app/reward/api/internal/rpc"
	"oshit-go/app/reward/api/internal/svc"
	"oshit-go/app/reward/api/types"
	app_utils "oshit-go/app/utils"
	"oshit-go/common/pkg/dal/model"
	"oshit-go/common/pkg/entity"
	"oshit-go/common/utils"
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
	inviteLogic       *logic.RewardInviteLogic
}

func NewTakeLogic(ctx context.Context, srvCtx *svc.ServiceContext) *TakeTokenLogic {
	return &TakeTokenLogic{
		prefix:            "TakeToken业务 -",
		ctx:               ctx,
		srvCtx:            srvCtx,
		db:                srvCtx.DB,
		rd:                srvCtx.Redis,
		rs:                srvCtx.RedSync,
		rpcClient:         srvCtx.RpcClient,
		baseClient:        srvCtx.BaseClient,
		levelDist:         srvCtx.LevelDist.DistLevel,
		levelRatio:        srvCtx.LevelRatio,
		levelRatioMap:     srvCtx.LevelRatioMap,
		serviceConfig:     srvCtx.TakeTokenConfig,
		decimals:          uint8(srvCtx.TokenConfig.Decimals),
		LightHouseAddress: srvCtx.LightHouseAddress,
		inviteLogic:       logic.NewRewardInviteLogic(ctx, srvCtx.DB),
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

	// 2. 递归向上查询需要奖励的邀请人，若邀请关系已确定则将直接邀请人置于列表首位
	var inviterForSort *model.NativeAccountInfo
	if (invited || codeValid) && directInviter != nil {
		inviterForSort = directInviter
	}
	sortedItems, sortedClaims, err := l.inviteLogic.BuildSortedInviterItems(
		receiptNativeAccount, l.levelDist, l.levelRatio, inviterForSort,
	)
	if err != nil {
		log.Errorf("%s 递归向上查询邀请人错误[%v]", prefix, err)
		return nil, errors.New("recursive query up inviter records error")
	}

	// 3. 根据是否确定邀请关系来决定奖励金额
	var rewardAmount uint64 = 0
	if !invited {
		rewardAmount = uint64(l.serviceConfig.Amount)
	} else {
		rewardAmount = uint64(l.serviceConfig.InviteAmount)
	}

	// 5. 填写领取奖励信息
	rewardInfo := types.RewardTokenItem{Index: 0, ReceiptAccount: receiptNativeAccount, Amount: rewardAmount}
	// 在最小集合里面决定每个层级的邀请人领取多少金额
	for index, claim := range sortedClaims {
		sortedItems[index].Amount = uint64(float64(rewardAmount) * claim.Ratio)
	}

	// 6. 获取本次奖励的总代币量
	totalRewardAmount := takeTxInfo.RewardInfo.Amount
	for _, item := range takeTxInfo.RewardInviterInfo {
		totalRewardAmount += item.Amount
	}

	// 7. 计算所有的token的价格
	quoteSOLPrice, err := l.baseClient.GetTokenQuoteSOLPrice(ctx)
	if err != nil {
		log.Errorf("%s 计算奖励金额价格错误: %v", prefix, err)
		return nil, fmt.Errorf("get toke quote sol price failed")
	}

	// 8. 填写最终需要返回的交易信息
	takeTxInfo.RewardAccount = l.serviceConfig.RewardAccount
	takeTxInfo.Mint = l.srvCtx.TokenConfig.Mint
	takeTxInfo.CostAccount = l.serviceConfig.CostAccount
	takeTxInfo.DexFeeRate = l.serviceConfig.DexFeeRate
	takeTxInfo.MaxDexFee = l.serviceConfig.MaxDexFee
	takeTxInfo.Decimals = int32(l.srvCtx.TokenConfig.Decimals)
	takeTxInfo.InviteCode = inviteCode
	takeTxInfo.RewardInfo = rewardInfo
	takeTxInfo.InviteCodeValid = codeValid
	takeTxInfo.QuoteSOLPrice = quoteSOLPrice
	takeTxInfo.TotalReward = float64(totalRewardAmount)
	takeTxInfo.QuotedSOLAmount = quoteSOLPrice * float64(totalRewardAmount) / l.srvCtx.TokenDecimal * float64(solana.LAMPORTS_PER_SOL)
	takeTxInfo.Invited = invited
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
	err = table.Where("receipt_account = ? and use_invite_code = ? and tx_state = ?", nativeAccount, true, 1).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, nil
}

// recordTakeToken 记录官网领取token记录
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
		RewardAccount:  decodedServiceTx.RewardInst.FromNativeAccount,
		ReceiptAccount: decodedServiceTx.RewardInst.ToNativeAccount,
		CostAccount:    decodedServiceTx.ToDexInst.ToNativeAccount,
		TxID:           decodedServiceTx.TxID,
		Amount:         l.serviceConfig.Amount,
		DexFee:         decodedServiceTx.ToDexInst.Amount,
		UseInviteCode:  takeTokenTxInfo.InviteCodeValid,
		InviteCode:     takeTokenTxInfo.InviteCode,
		TxState:        0,
		Invited:        invited,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
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

// checkNeedLottery 检查用户是否需要完成抽奖才能继续领取
func (l *TakeTokenLogic) checkNeedLottery(ctx context.Context, nativeAccount string) error {
	today := time.Now().Truncate(24 * time.Hour)
	var stats model.DailyClaimStats
	err := l.db.Table(model.TableNameDailyClaimStats).
		Where("native_account = ? and take_date = ?", nativeAccount, today).
		First(&stats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 今天没有记录，不需要抽奖
			return nil
		}
		return err
	}
	if stats.NeedLottery {
		return errors.New("you must complete lottery before taking token again")
	}
	return nil
}

// ProcessCommitTx 处理提交上来的交易
func (l *TakeTokenLogic) ProcessCommitTx(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx, inviteCode string) error {
	txIdStr := preCheckedTx.TxId.String()
	prefix := fmt.Sprintf("%s 处理用户提交交易Id %v -", l.prefix, txIdStr)

	// 1. 分布式锁，防止重复处理同一地址短时间内重复进行业务
	mutex := l.rs.NewMutex("take-token:process:commit-tx:"+preCheckedTx.From.String(), redsync.WithExpiry(1*time.Hour))
	if err := mutex.Lock(); err != nil {
		var errTaken *redsync.ErrTaken
		if !errors.As(err, &errTaken) {
			log.Errorf("%s 获取处理交易的分布式锁错误: %v", prefix, err)
			return errors.New("process transaction error")
		}
		return errors.New("you have unfinished service.please try again later")
	}
	defer mutex.Unlock()

	// 2. 检查是否必须要抽奖
	if err := l.checkNeedLottery(ctx, preCheckedTx.From.String()); err != nil {
		log.Errorf("%s 获取钱包是否需要抽奖错误: %v", prefix, err)
		return err
	}
	// 3. 获取take token的交易信息
	takeTxInfo, err := l.getTxInfo(ctx, preCheckedTx.From.String(), inviteCode)
	if err != nil {
		log.Errorf("%s 获取交易信息错误: %v", prefix, err)
		return fmt.Errorf("get take token transaction info error: %v", err)
	}

	// 4. 解析出来交易里面的 transfer checked 和 transfer 指令集合
	decodedTx, err := l.decodeSOLTx(takeTxInfo, &preCheckedTx.SOLTx)
	if err != nil {
		log.Errorf("%s 解析solana交易错误: %v", prefix, err)
		return fmt.Errorf("decode transaction error:%s", err)
	}

	// 5. 检查 transfer checked 指令和 transfer 指令是否符合奖励要求
	decodedServiceTx, err := l.checkDecodedSOLTx(takeTxInfo, decodedTx)
	if err != nil {
		log.Errorf("%s 校验solana交易当中的指令错误: %v", prefix, err)
		return fmt.Errorf("check transaction instruction failed: %v", err)
	}

	// 6. 通过 base 模块的dubbo接口签名并异步广播
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Reward", "TakeToken")
	if err != nil {
		log.Errorf("%s 调用base模块dubbo接口发送交易失败,错误: %v", prefix, err)
		return errors.New(utils.FilterAndTranslateSOLError(err))
	}

	// 7. 如果发送的交易Id与预期的不一致，则报错
	if sentTxId == "" || sentTxId != txIdStr {
		log.Errorf("%s 调用base模块dubbo接口发送的交易Id %s 与预期的不一致", prefix, sentTxId)
		return errors.New("sent transaction id not equal expected")
	}

	// 8. 记录领取记录到数据库
	decodedServiceTx.TxID = txIdStr
	if err = l.recordTakeToken(takeTxInfo, decodedServiceTx, takeTxInfo.Invited); err != nil {
		log.Errorf("%s 记录交易信息,错误: %v", prefix, err)
		return errors.New("record official transfer token error")
	}

	return nil
}
