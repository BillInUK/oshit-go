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
	mint          solana.PublicKey
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

func (l *GiveTokenLogic) GetTxInfo(ctx context.Context, from, to string, amountUI float64) (*types.GiveTokenTxInfo, error) {
	//levelDist := l.srvCtx.LevelDist.Level
	//levelRatio := l.srvCtx.LevelRatio
	//// 转换金额
	//amoutInst := uint64(amountUI * l.srvCtx.TokenDecimal)
	//// 获取token兑换sol的价格
	//quoteSOLPrice := l.baseClient.GetTokenQuoteSOLPrice(ctx)
	//// 先查询上级邀请人
	//upInviters, err := l.baseClient.GetUpInviterRecords(ctx, from, levelDist)
	//if err != nil {
	//	log.Errorf("%s 查询地址 %s 上级邀请人错误: %v", l.prefix, from, err)
	//	return nil, fmt.Errorf("get from up inviters error: %v", err)
	//}
	//inviteRecords := upInviters.Records
	//// 检查地址是否是有效地址
	//valid := l.accountValid(to)
	//
	//// 根据奖励规则生成指令
	//rewardTokenItem := RewardTokenItem
	// 计算总的奖励金额

	// 计算成本费

	// 返回最终的指令
	//type GiveTokenTxInfo struct {
	//	RewardNativeAccount string  `json:"rewardNativeAccount"`
	//	RewardTokenAccount  string  `json:"rewardTokenAccount"`
	//	TokenMintAccount    string  `json:"tokenMintAccount"`
	//	DexNativeAccount    string  `json:"dexAccount"`
	//	DexFeeRate          float64 `json:"dexFeeRate"`
	//	MaxDexFee           float64 `json:"maxDexFee"`
	//	Decimals            int32   `json:"decimals"`
	//	QuoteSOLPrice       float64 `json:"quoteSOLPrice"`
	//
	//	TotalRewardAmount float64            `json:"totalRewardAmount"`
	//	QuotedSOLAmount   float64            `json:"quotedSOLAmount"`
	//	Claims            []model.LevelRatio `json:"claims"`
	//	RewardInfo        RewardTokenItem    `json:"rewardInfo"`
	//	RewardInviterInfo []RewardTokenItem  `json:"rewardInviterInfo"`
	//}
	txInfo := &types.GiveTokenTxInfo{
		RewardNativeAccount: l.serviceConfig.RewardNativeAccount,
		RewardTokenAccount:  l.serviceConfig.RewardTokenAccount,
		TokenMintAccount:    l.serviceConfig.TokenMintAccount,
		DexNativeAccount:    l.serviceConfig.DexNativeAccount,
		DexFeeRate:          l.serviceConfig.RewardRate,
		MaxDexFee:           l.serviceConfig.MaxValidReward,
		Decimals:            l.serviceConfig.Decimal,
	}

	return txInfo, nil
}

// getLatestRecord 查询最新1笔记录
func (l *GiveTokenLogic) getLatestRecord(nativeAccount string) (*model.GiveTokenRecord, error) {
	var record model.GiveTokenRecord
	table := l.db.Table(model.TableNameGiveTokenRecord)
	err := table.Where("from_native_account = ?", nativeAccount).Order("created_at desc").First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &record, nil
}

// getDailyRecords 查询当日的领取记录
func (l *GiveTokenLogic) getDailyRecords(nativeAccount string) ([]model.GiveTokenRecord, error) {
	var records []model.GiveTokenRecord
	table := l.db.Table(model.TableNameGiveTokenRecord)
	// 将当前时间的时间部分截断，只保留日期部分
	currentDate := time.Now().Truncate(24 * time.Hour)
	// 查询当天的记录
	if err := table.Where(" from_native_account =  ? and created_at >= ? AND created_at < ?",
		nativeAccount, currentDate, currentDate.Add(24*time.Hour)).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// takeTokenRecordExist 是否有过领取记录正在领取或者领取成功的
func (l *GiveTokenLogic) takeTokenRecordExist(nativeAccount string) (bool, error) {
	var record model.TakeTokenRecord
	table := l.db.Table(model.TableNameTakeTokenRecord)
	if err := table.Where("receipt_native_account = ? and state >= 0", nativeAccount).First(&record).Error; err != nil {
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
	if err := table.Where("receipt_native_account = ? and state >= 0", nativeAccount).First(&record).Error; err != nil {
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
	// 查看有没有过token的转账记录
	if exist, err := l.takeTokenRecordExist(nativeAccount.String()); err != nil || exist {
		return false, err
	}
	// 查看是否有过官网转账记录
	if exist, err := l.giveTokenRecordExist(nativeAccount.String()); err != nil || exist {
		return false, err
	}
	// 如果都没有记录，则再判断token account是否存在
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
		// 如果token account不存在则是有效地址
		if err.Error() == "not found" {
			return true, nil
		}
		return false, nil
	}
	return false, nil
}

// sendTransaction 签名并广播用户提交上来的交易
func (l *GiveTokenLogic) sendTransaction(ctx context.Context, tx *solana.Transaction, service string) (*solana.Signature, error) {
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

// recordGiveToken 记录官网转token记录
func (l *GiveTokenLogic) recordGiveToken(serviceTx *entity.DecodedServiceTransaction) error {
	// 开启事务
	dbTx := l.db.Begin()
	if dbTx.Error != nil {
		log.Errorf("官网转账流程 - 开启事务错误: %v", dbTx.Error)
		return dbTx.Error
	}

	// 记录领取的交易ID
	txRecord := model.ServiceTx{
		Service:    "Reward",
		SubService: "GiveToken",
		TxID:       serviceTx.TxID,
		CreatedAt:  time.Now(),
	}
	if err := dbTx.Table(model.TableNameServiceTx).Create(&txRecord).Error; err != nil {
		dbTx.Rollback() // 回滚事务
		log.Errorf("官网转账流程 - 插入交易业务类型表错误: %v", err)
		return err
	}

	inst := serviceTx.TransferTokenInst
	record := model.GiveTokenRecord{
		FromNativeAccount:    serviceTx.FromNativeAccount,
		FromTokenAccount:     serviceTx.FromTokenAccount,
		ReceiptTokenAccount:  inst.ToTokenAccount,
		ReceiptNativeAccount: inst.ToNativeAccount,
		TxID:                 serviceTx.TxID,
		Amount:               inst.Amount,
		State:                0,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	if err := dbTx.Table(model.TableNameGiveTokenRecord).Create(&record).Error; err != nil {
		dbTx.Rollback() // 回滚事务
		log.Errorf("官网转账流程 - 插入转账记录表错误: %v", err)
		return err
	}

	// 提交事务
	if err := dbTx.Commit().Error; err != nil {
		log.Errorf("官网转账流程 - 提交事务错误: %v", err)
		return err
	}

	return nil
}

func (l *GiveTokenLogic) limitExceed() (bool, error) {
	// 查询最近是否有过领取奖励记录
	//transferRecord, err := l.getLatestRecord(preCheckedTx.From.String())
	//if err != nil {
	//	log.Errorf("官方转账获取奖励 - 查询领取奖励记录错误:%v", err)
	//	return nil,errors.New("query official give token record error")
	//}
	//if transferRecord != nil {
	//	// 同1地址两次领取之间不能超过5分钟
	//	if isTimeDifferenceLessThan(transferRecord.CreatedAt, removeTz(time.Now()), time.Duration(serviceConfig.Interval)*time.Minute) {
	//		log.Errorf("官方转账获取奖励 - 转账过于频繁")
	//		return nil,fmt.Errorf("transfer is too frequent. please try again after %d minutes", serviceConfig.Interval)
	//	}
	//	// 同1地址1天内领取次数不能超过20次
	//	recentRecords, err := l.getDailyRecords(preCheckedTx.From.String())
	//	if err != nil {
	//		log.Errorf("官方转账获取奖励 - 查询当日转账记录记录错误: %v", err)
	//		return nil, errors.New("query daily transfer token records error")
	//	}
	//	if len(recentRecords) >= 20 {
	//		log.Errorf("官方转账获取奖励 - 地址 %v 在当日转账的次数超过限制: %v", preCheckedTx.From.String(), err)
	//		return nil,errors.New("limit exceeded: Max 20 transfer per wallet per day."))
	//	}
	//}
	return true, nil
}

// getUpInviters 查询上级邀请人的以及每个上级邀请人所能拿到的奖励费率
func (l *GiveTokenLogic) getUpInviters(nativeAccount string) ([]model.InviteRelation, error) {
	inviteRecords, err := logic.NewRewardInviteLogic(l.ctx, l.srvCtx.DB).GetUpInviterRecords(nativeAccount, l.srvCtx.LevelDist.Level)
	if err != nil {
		return nil, err
	}
	// 获取两个数组的最小长度
	minLength := min(len(l.srvCtx.LevelRatio), len(inviteRecords))
	// 返回截取后的数组
	return inviteRecords[:minLength], nil
}

// ProcessCommitTx 处理前端提交的 give token 交易
func (l *GiveTokenLogic) ProcessCommitTx(ctx context.Context, preCheckedTx *app_utils.PreCheckedTx, toNativeAccount solana.PublicKey) (*solana.Signature, error) {
	prefix := fmt.Sprintf("%s - %s -", l.prefix, "处理钱包提交交易")

	// TODO: 调用 limitExceed 检查是否超过限制

	// 分布式锁，防止重复的对有效地址转账，重复领取有效地址奖励
	mutex := l.rs.NewMutex("give-token:process:commit-tx:"+preCheckedTx.From.String(), redsync.WithExpiry(1*time.Hour))
	if err := mutex.Lock(); err != nil {
		var errTaken *redsync.ErrTaken
		if !errors.As(err, &errTaken) {
			log.Errorf("%s 获取处理交易的分布式锁错误: %v", prefix, err)
			return nil, errors.New("process transaction error")
		}
		return nil, errors.New("you have unfinished transfer.please try again later")
	}
	defer mutex.Unlock()

	// 查询地址是否是有效地址
	valid, err := l.accountValid(toNativeAccount.String())
	if err != nil {
		log.Errorf("%s 查询地址是否是有效地址错误:[%v]", prefix, err)
		return nil, errors.New("query receipt account match reward rule error")
	}
	// 查询转账用户的邀请上级以及每层上级领取的费率
	upInvitersInfo, err := l.getUpInviters(preCheckedTx.From.String())
	if err != nil {
		log.Errorf("%s 查找转账地址[%v]的邀请上级和邀请上级奖励费率失败", prefix, preCheckedTx.From.String())
		return nil, errors.New("query claims and inviter error")
	}
	tokenMintAccount, _ := solana.PublicKeyFromBase58(l.serviceConfig.TokenMintAccount)
	transferToTokenAccount, _, _ := solana.FindAssociatedTokenAddress(toNativeAccount, tokenMintAccount)
	// 解析出来交易里面的transfer checked和transfer指令集合
	// TODO: 考虑一下是否需要创建 token account的情况
	decodedTx, err := l.decodeSOLTx(l.rpcClient, tokenMintAccount, toNativeAccount, true, &preCheckedTx.SOLTx)
	if err != nil {
		log.Errorf("%s 解析solana交易失败，错误: %v", prefix, err)
		return nil, errors.New("decode solana transaction error")
	}
	// 检查transfer checked指令和transfer指令是否符合奖励要求
	decodedServiceTx, err := l.checkSOLTx(decodedTx, upInvitersInfo, transferToTokenAccount, valid)
	if err != nil {
		log.Errorf("%s 检查交易当中的指令失败，错误: %v", prefix, err)
		return nil, err
	}
	// 实际发送交易
	go func() {
		// TODO: 异步的调用dubbo发送交易
		if _, err := l.sendTransaction(ctx, &preCheckedTx.SOLTx, "GiveToken"); err != nil {
			log.Errorf("%s 签名并发送交易错误: %v", prefix, err)
		}
	}()

	// 记录交易信息到数据库
	decodedServiceTx.TxID = preCheckedTx.TxId.String()
	if err := l.recordGiveToken(decodedServiceTx); err != nil {
		log.Errorf("%s 记录交易信息到数据库错误: %v", prefix, err)
		return nil, errors.New("record give token error")
	}

	return &preCheckedTx.TxId, nil
}
