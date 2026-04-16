package give

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

type GiveTokenLogic struct {
	prefix        string
	ctx           context.Context
	srvCtx        *svc.ServiceContext
	db            *gorm.DB
	rd            redis.UniversalClient
	rs            redsync.Redsync
	rpcClient     *rpc.Client
	baseClient    *rewardrpc.BaseClient
	serviceConfig *model.GiveTokenConfig
	inviteLogic   *logic.RewardInviteLogic
}

func NewGiveTokenLogic(ctx context.Context, srvCtx *svc.ServiceContext) *GiveTokenLogic {
	return &GiveTokenLogic{
		prefix:        "GiveToken",
		ctx:           ctx,
		srvCtx:        srvCtx,
		db:            srvCtx.DB,
		rd:            srvCtx.Redis,
		rs:            srvCtx.RedSync,
		rpcClient:     srvCtx.RpcClient,
		baseClient:    srvCtx.BaseClient,
		serviceConfig: srvCtx.GiveTokenConfig,
		inviteLogic:   logic.NewRewardInviteLogic(ctx, srvCtx.DB),
	}
}

func (l *GiveTokenLogic) GetDefaultConfig() (*model.GiveTokenConfig, error) {
	var config model.GiveTokenConfig
	table := l.db.Table(model.TableNameGiveTokenConfig)
	if err := table.First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

func (l *GiveTokenLogic) GetRecord(ctx context.Context, txId string) (*model.GiveTokenRecord, error) {
	var record model.GiveTokenRecord
	table := l.db.Table(model.TableNameGiveTokenRecord)
	if err := table.Where("tx_id = ?", txId).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// GetTxInfo 获取 give token 的交易参数
func (l *GiveTokenLogic) GetTxInfo(ctx context.Context, from, to string, amountUI float64) (*types.GiveTokenTxInfo, error) {
	prefix := fmt.Sprintf("%s GetTxInfo from=%s to=%s -", l.prefix, from, to)

	// 1. 确定 to 地址的 token account 是否存在，决定奖励费率
	toNativeKey, _ := solana.PublicKeyFromBase58(to)
	tokenMintKey, _ := solana.PublicKeyFromBase58(l.srvCtx.TokenConfig.Mint)
	toTokenAccount, _, _ := solana.FindAssociatedTokenAddress(toNativeKey, tokenMintKey)
	accountInfo, _ := l.rpcClient.GetAccountInfo(ctx, toTokenAccount)
	toTokenAccountExists := accountInfo != nil && accountInfo.Value != nil

	// 2. 计算 from 的奖励金额（原始单位）
	amountRaw := uint64(amountUI * l.srvCtx.TokenDecimal)
	var rewardAmount float64
	if toTokenAccountExists {
		rewardAmount = min(l.serviceConfig.MaxValidReward, float64(amountRaw)*l.serviceConfig.RewardRate/100)
	} else {
		rewardAmount = min(l.serviceConfig.MaxValidReward, float64(amountRaw)*l.serviceConfig.ValidRate/100)
	}

	// 3. 递归向上查询需要奖励的邀请人
	sortedItems, sortedClaims, err := l.inviteLogic.BuildSortedInviterItems(
		from, l.srvCtx.LevelDist.DistLevel, l.srvCtx.LevelRatio, nil,
	)
	if err != nil {
		log.Errorf("%s 递归向上查询邀请人错误: %v", prefix, err)
		return nil, fmt.Errorf("recursive query up inviter records error")
	}

	// 4. 确定每个邀请人的奖励金额
	for index, claim := range sortedClaims {
		sortedItems[index].Amount = uint64(rewardAmount * claim.Ratio / 100)
	}

	// 5. 计算总奖励金额
	totalReward := rewardAmount
	for _, item := range sortedItems {
		totalReward += float64(item.Amount)
	}

	// 6. 获取 token/SOL 价格
	quoteSOLPrice, err := l.baseClient.GetTokenQuoteSOLPrice(ctx)
	if err != nil {
		log.Errorf("%s 获取 token/SOL 价格失败: %v", prefix, err)
		return nil, fmt.Errorf("get token quote sol price failed")
	}

	// 7. 构造返回值
	giveInfo := types.RewardTokenItem{Index: 0, ReceiptAccount: to, Amount: amountRaw}
	rewardInfo := types.RewardTokenItem{Index: 0, ReceiptAccount: from, Amount: uint64(rewardAmount)}

	txInfo := &types.GiveTokenTxInfo{
		RewardAccount:     l.serviceConfig.RewardAccount,
		Mint:              l.srvCtx.TokenConfig.Mint,
		CostAccount:       l.serviceConfig.CostAccount,
		DexFeeRate:        l.serviceConfig.RewardRate,
		MaxDexFee:         l.serviceConfig.MaxValidReward,
		Decimals:          int32(l.srvCtx.TokenConfig.Decimals),
		QuoteSOLPrice:     quoteSOLPrice,
		TotalReward:       totalReward,
		QuotedSOLAmount:   quoteSOLPrice * totalReward / l.srvCtx.TokenDecimal * float64(solana.LAMPORTS_PER_SOL),
		GiveInfo:          giveInfo,
		RewardInfo:        rewardInfo,
		Claims:            sortedClaims,
		RewardInviterInfo: sortedItems,
	}

	return txInfo, nil
}

// getLatestRecord 查询最新1笔记录
func (l *GiveTokenLogic) getLatestRecord(nativeAccount string) (*model.GiveTokenRecord, error) {
	var record model.GiveTokenRecord
	table := l.db.Table(model.TableNameGiveTokenRecord)
	err := table.Where("from_account = ?", nativeAccount).Order("created_at desc").First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, nil
}

// getDailyRecords 查询当日的领取记录
func (l *GiveTokenLogic) getDailyRecords(nativeAccount string) ([]model.GiveTokenRecord, error) {
	var records []model.GiveTokenRecord
	table := l.db.Table(model.TableNameGiveTokenRecord)
	currentDate := time.Now().Truncate(24 * time.Hour)
	if err := table.Where(" from_account =  ? and created_at >= ? AND created_at < ?",
		nativeAccount, currentDate, currentDate.Add(24*time.Hour)).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// takeTokenRecordExist 是否有过领取记录正在领取或者领取成功的
func (l *GiveTokenLogic) takeTokenRecordExist(nativeAccount string) (bool, error) {
	var record model.TakeTokenRecord
	table := l.db.Table(model.TableNameTakeTokenRecord)
	if err := table.Where("receipt_account = ? and tx_state >= 0", nativeAccount).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// giveTokenRecordExist 是否有过领取记录正在领取或者领取成功的
func (l *GiveTokenLogic) giveTokenRecordExist(nativeAccount string) (bool, error) {
	var record model.GiveTokenRecord
	table := l.db.Table(model.TableNameGiveTokenRecord)
	if err := table.Where("receipt_account = ? and tx_state >= 0", nativeAccount).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// accountValid 查看地址是否是有效地址
func (l *GiveTokenLogic) accountValid(account string) (bool, error) {
	nativeAccount, err := solana.PublicKeyFromBase58(account)
	if err != nil {
		return false, errors.New("can not convert account to solana public key")
	}
	if exist, err := l.takeTokenRecordExist(nativeAccount.String()); err != nil || exist {
		return false, err
	}
	if exist, err := l.giveTokenRecordExist(nativeAccount.String()); err != nil || exist {
		return false, err
	}
	tokenMintAccount, err := solana.PublicKeyFromBase58(l.srvCtx.TokenConfig.Mint)
	if err != nil {
		return false, err
	}
	tokenAccount, _, _ := solana.FindAssociatedTokenAddress(nativeAccount, tokenMintAccount)
	_, err = l.rpcClient.GetAccountInfoWithOpts(
		context.Background(),
		tokenAccount,
		&rpc.GetAccountInfoOpts{
			Encoding: solana.EncodingJSONParsed,
		},
	)
	if err != nil {
		if err.Error() == "not found" {
			return true, nil
		}
		return false, nil
	}
	return false, nil
}

// getUpInviters 查询上级邀请人的以及每个上级邀请人所能拿到的奖励费率
func (l *GiveTokenLogic) getUpInviters(nativeAccount string) ([]model.InviteRelation, error) {
	inviteRecords, err := l.inviteLogic.GetUpInviterRecords(nativeAccount, l.srvCtx.LevelDist.DistLevel)
	if err != nil {
		return nil, err
	}
	minLength := min(len(l.srvCtx.LevelRatio), len(inviteRecords))
	return inviteRecords[:minLength], nil
}

// recordGiveToken 记录官网转token记录
func (l *GiveTokenLogic) recordGiveToken(serviceTx *entity.DecodedServiceTransaction) error {
	dbTx := l.db.Begin()
	if dbTx.Error != nil {
		log.Errorf("官网转账流程 - 开启事务错误: %v", dbTx.Error)
		return dbTx.Error
	}

	txRecord := model.ServiceTx{
		Service:    "Reward",
		SubService: "GiveToken",
		TxID:       serviceTx.TxID,
		CreatedAt:  time.Now(),
	}
	if err := dbTx.Table(model.TableNameServiceTx).Create(&txRecord).Error; err != nil {
		dbTx.Rollback()
		log.Errorf("官网转账流程 - 插入交易业务类型表错误: %v", err)
		return err
	}

	inst := serviceTx.TransferTokenInst
	record := model.GiveTokenRecord{
		FromAccount:    serviceTx.FromNativeAccount,
		ReceiptAccount: inst.ToNativeAccount,
		TxID:           serviceTx.TxID,
		Amount:         inst.Amount,
		TxState:        0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := dbTx.Table(model.TableNameGiveTokenRecord).Create(&record).Error; err != nil {
		dbTx.Rollback()
		log.Errorf("官网转账流程 - 插入转账记录表错误: %v", err)
		return err
	}

	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("官网转账流程 - 提交事务错误: %v", err)
		return err
	}
	return nil
}

func (l *GiveTokenLogic) limitExceed() (bool, error) {
	return true, nil
}

// ProcessCommitTx 处理前端提交的 give token 交易
func (l *GiveTokenLogic) ProcessCommitTx(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx, toNativeAccount solana.PublicKey) error {
	txIdStr := preCheckedTx.TxId.String()
	prefix := fmt.Sprintf("%s 处理用户提交交易Id %v -", l.prefix, txIdStr)

	// 1. 分布式锁，防止重复处理同一地址短时间内重复进行业务
	mutex := l.rs.NewMutex("give-token:process:commit-tx:"+preCheckedTx.From.String(), redsync.WithExpiry(1*time.Hour))
	if err := mutex.Lock(); err != nil {
		var errTaken *redsync.ErrTaken
		if !errors.As(err, &errTaken) {
			log.Errorf("%s 获取处理交易的分布式锁错误: %v", prefix, err)
			return errors.New("process transaction error")
		}
		return errors.New("you have unfinished service.please try again later")
	}
	defer mutex.Unlock()

	// 2. 查询地址是否是有效地址
	valid, err := l.accountValid(toNativeAccount.String())
	if err != nil {
		log.Errorf("%s 查询接收地址是否是有效地址错误: %v", prefix, err)
		return errors.New("query receipt account match reward rule error")
	}

	// 3. 查询转账用户的邀请上级
	upInvitersInfo, err := l.getUpInviters(preCheckedTx.From.String())
	if err != nil {
		log.Errorf("%s 查找转账地址上级邀请人错误: %v", prefix, err)
		return errors.New("query claims and inviter error")
	}

	tokenMintAccount, _ := solana.PublicKeyFromBase58(l.srvCtx.TokenConfig.Mint)
	transferToTokenAccount, _, _ := solana.FindAssociatedTokenAddress(toNativeAccount, tokenMintAccount)

	// 4. 解析 solana 交易
	decodedTx, err := l.decodeSOLTx(l.rpcClient, tokenMintAccount, toNativeAccount, true, &preCheckedTx.SOLTx)
	if err != nil {
		log.Errorf("%s 解析solana交易失败错误: %v", prefix, err)
		return errors.New("decode solana transaction error")
	}

	// 5. 检查交易指令是否符合规则,并返回业务交易
	decodedServiceTx, err := l.checkSOLTx(decodedTx, upInvitersInfo, transferToTokenAccount, valid)
	if err != nil {
		log.Errorf("%s 检查交易当中的指令失败错误: %v", prefix, err)
		return err
	}

	// 6. 通过 base 模块的dubbo接口签名并异步广播
	sentTxId, err := l.baseClient.SendTransaction(ctx, &preCheckedTx.SOLTx, "Reward", "GiveToken")
	if err != nil {
		log.Errorf("%s 调用base模块dubbo接口发送交易失败,错误: %v", prefix, err)
		return errors.New(utils.FilterAndTranslateSOLError(err))
	}
	// 如果发送的交易Id与预期的不一致，则报错
	if sentTxId == "" || sentTxId != txIdStr {
		log.Errorf("%s 调用base模块dubbo接口发送的交易Id %s 与预期的不一致", prefix, sentTxId)
		return errors.New("sent transaction id not equal expected")
	}

	// 7. 记录领取记录到数据库
	txId := preCheckedTx.SOLTx.Signatures[0]
	decodedServiceTx.TxID = txId.String()
	if err = l.recordGiveToken(decodedServiceTx); err != nil {
		log.Errorf("%s 记录交易信息,错误: %v", prefix, err)
		return errors.New("record give token error")
	}

	return nil
}
