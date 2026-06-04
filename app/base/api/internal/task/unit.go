package task

import (
	"context"
	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/programs/tokenregistry"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/go-redsync/redsync/v4"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"oshit-go/common/pkg/dal/model"
	"time"
)

// 分布式锁 key
const unitEstimateLock = "base:sol:unit:estimate:lock"

// Redis 数据 key
const (
	unitAssociatedAccountRent = "base:sol:unit:associated-account-rent"
	unitAssociatedAccount     = "base:sol:unit:associated-account"
	unitTransferChecked       = "base:sol:unit:transfer-checked"
	unitMemo                  = "base:sol:unit:memo"
)

// UnitTask 手续费统计任务
type UnitTask struct {
	db          *gorm.DB
	redis       redis.UniversalClient
	redSync     redsync.Redsync
	rpcClient   *rpc.Client
	rpcURL      string
	tokenConfig *model.TokenConfig
	// 预构建的模拟交易（静态数据，只需初始化一次）
	fromNativeAccount2     solana.PublicKey
	computeBudgetPriceInst solana.Instruction
	refBlockHash           solana.Hash
	sig1                   solana.Signature
	sig2                   solana.Signature
	transferCheckedTx      *solana.Transaction
	memoTx                 *solana.Transaction
}

// NewUnitTask 创建手续费任务
func NewUnitTask(taskCtx *TaskContext) *UnitTask {
	var rpcClient *rpc.Client
	var rpcURL string
	if taskCtx.RpcPool != nil {
		rpcClient = taskCtx.RpcPool.First()
		rpcURL = taskCtx.RpcPool.FirstURL()
	} else if taskCtx.RpcClient != nil {
		rpcClient = taskCtx.RpcClient
	}
	return &UnitTask{
		db:          taskCtx.DB,
		redis:       taskCtx.Redis,
		redSync:     taskCtx.RedSync,
		rpcClient:   rpcClient,
		rpcURL:      rpcURL,
		tokenConfig: taskCtx.TokenConfig,
	}
}

// Start 启动任务
func (t *UnitTask) Start() {
	t.init()
	go runPeriodic(&t.redSync, 10*time.Second, unitEstimateLock, 5*time.Minute, t.simulateUnitConsumed)
}

// init 初始化模拟交易所需的静态数据
func (t *UnitTask) init() {
	t.refBlockHash, _ = solana.HashFromBase58("2yPEzFPcuKzbL9hG8QoStCAsQmrkrcpxMLTi7RM5qAgd")
	t.sig1, _ = solana.SignatureFromBase58("5iXm4LUMDHwtMTXvJjHgWmG41d5tFSmRn94jvzntxjd193ryZdpoWuYBXmgKD3KVw41uycsqAHH92VcU53rpDPJk")
	t.sig2, _ = solana.SignatureFromBase58("2EdJBKNNfodAsy6nFasWqYMtgFgU6rgY8WeckBMRHtgBEfdsn2pnKJfcj4LKABf76UonsTtn8UsPn7ZcbXMXkB5U")

	fromNativeAccount1, _ := solana.PublicKeyFromBase58("GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E")
	t.fromNativeAccount2, _ = solana.PublicKeyFromBase58("5D4MWh35wxUcY1hBsm5GwuippPL2UBmfDnfkC8MeqxcN")
	tokenMintAccount, _ := solana.PublicKeyFromBase58(t.tokenConfig.Mint)
	toNativeAccount1, _ := solana.PublicKeyFromBase58("27htRMGeQ4HV32SPHsJrpndZn1zwmF2kiPqmABcHgehx")
	fromTokenAccount, _, _ := solana.FindAssociatedTokenAddress(fromNativeAccount1, tokenMintAccount)
	toTokenAccount1, _, _ := solana.FindAssociatedTokenAddress(toNativeAccount1, tokenMintAccount)

	t.computeBudgetPriceInst = computebudget.NewSetComputeUnitPriceInstructionBuilder().SetMicroLamports(1000).Build()
	transferCheckedInst := token.NewTransferCheckedInstructionBuilder().
		SetAmount(1).
		SetDecimals(uint8(t.tokenConfig.Decimals)).
		SetSourceAccount(fromTokenAccount).
		SetMintAccount(tokenMintAccount).
		SetDestinationAccount(toTokenAccount1).
		SetOwnerAccount(fromNativeAccount1).
		Build()
	transferInst := system.NewTransferInstructionBuilder().
		SetLamports(1).
		SetFundingAccount(t.fromNativeAccount2).
		SetRecipientAccount(fromNativeAccount1).
		Build()

	t.transferCheckedTx, _ = solana.NewTransaction(
		[]solana.Instruction{t.computeBudgetPriceInst, transferCheckedInst, transferInst},
		t.refBlockHash,
		solana.TransactionPayer(fromNativeAccount1),
	)
	t.transferCheckedTx.Signatures = append(t.transferCheckedTx.Signatures, t.sig1, t.sig2)

	memoInst := solana.NewInstruction(
		solana.MemoProgramID,
		[]*solana.AccountMeta{solana.Meta(fromNativeAccount1).SIGNER().WRITE()},
		[]byte("hello!!!"),
	)
	t.memoTx, _ = solana.NewTransaction(
		[]solana.Instruction{memoInst, transferInst},
		t.refBlockHash,
		solana.TransactionPayer(fromNativeAccount1),
	)
	t.memoTx.Signatures = append(t.memoTx.Signatures, t.sig1, t.sig2)
}

// simulateUnitConsumed 评估常用指令的compute unit
func (t *UnitTask) simulateUnitConsumed() {
	miniRent, err := t.rpcClient.GetMinimumBalanceForRentExemption(
		context.Background(),
		tokenregistry.TOKEN_META_SIZE,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		log.Errorf("solana周期性评估计算单元 获取账户最小租金错误: %v", err)
		return
	}

	if err := t.redis.Set(context.Background(), unitAssociatedAccountRent, miniRent, 1*time.Hour).Err(); err != nil {
		log.Errorf("solana周期性评估 - 更新最小账户租金到 - 刷新Redis错误: %v", err)
		return
	}

	createAccountInstruction := system.NewCreateAccountInstruction(
		miniRent,
		tokenregistry.TOKEN_META_SIZE,
		tokenregistry.ProgramID(),
		t.fromNativeAccount2,
		solana.NewWallet().PublicKey(),
	).Build()
	associatedTx, _ := solana.NewTransaction(
		[]solana.Instruction{t.computeBudgetPriceInst, createAccountInstruction},
		t.refBlockHash,
		solana.TransactionPayer(t.fromNativeAccount2),
	)
	associatedTx.Signatures = append(associatedTx.Signatures, t.sig1, t.sig2)

	srAssociateTokenAccount, err := t.rpcClient.SimulateTransactionWithOpts(context.Background(), associatedTx, &rpc.SimulateTransactionOpts{ReplaceRecentBlockhash: true})
	if err != nil {
		log.Errorf("solana周期性评估 associate token account 的Compute Unit错误: %v", err)
		return
	}
	if err := t.redis.Set(context.Background(), unitAssociatedAccount, *srAssociateTokenAccount.Value.UnitsConsumed, 1*time.Hour).Err(); err != nil {
		log.Errorf("solana周期性更新优先费用，刷新Redis错误: %v", err)
		return
	}

	srTransferChecked, err := t.rpcClient.SimulateTransactionWithOpts(context.Background(), t.transferCheckedTx, &rpc.SimulateTransactionOpts{ReplaceRecentBlockhash: true})
	if err != nil {
		log.Errorf("solana周期性评估 compute unit 1 错误: %v", err)
		return
	}
	log.Debugf("solana周期性更新优先费用 获取transfer checked unit consumed %d", *srTransferChecked.Value.UnitsConsumed)
	if err := t.redis.Set(context.Background(), unitTransferChecked, *srTransferChecked.Value.UnitsConsumed, 1*time.Hour).Err(); err != nil {
		log.Errorf("solana周期性更新优先费用，刷新Redis错误: %v", err)
		return
	}

	srMemo, err := t.rpcClient.SimulateTransactionWithOpts(context.Background(), t.memoTx, &rpc.SimulateTransactionOpts{ReplaceRecentBlockhash: true})
	if err != nil {
		log.Errorf("solana周期性评估 memo 错误: %v", err)
		return
	}
	if err := t.redis.Set(context.Background(), unitMemo, *srMemo.Value.UnitsConsumed, 1*time.Hour).Err(); err != nil {
		log.Errorf("solana周期性更新优先费用，刷新Redis错误: %v", err)
		return
	}
}
