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
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"oshit-go/app/base/dal/model"
	"time"
)

// UnitTask 手续费统计任务
type UnitTask struct {
	db          *gorm.DB
	redis       redis.UniversalClient
	redSync     redsync.Redsync
	rpcClient   *rpc.Client
	rpcURL      string
	tokenConfig *model.TokenConfig
}

// NewUnitTask 创建手续费任务
func NewUnitTask(taskCtx *TaskContext) *UnitTask {
	rpcURL := taskCtx.ChainConfig.RPCURL
	rpcClient := rpc.New(rpcURL)
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
	go t.startSimulateUnitConsumed()
}

// startSimulateUnitConsumed 周期性的评估常用指令的compute unit
func (t *UnitTask) startSimulateUnitConsumed() {
	var transferCheckedInsts []solana.Instruction
	var memoInsts []solana.Instruction
	refBlockHash, _ := solana.HashFromBase58("2yPEzFPcuKzbL9hG8QoStCAsQmrkrcpxMLTi7RM5qAgd")
	sig1, _ := solana.SignatureFromBase58("5iXm4LUMDHwtMTXvJjHgWmG41d5tFSmRn94jvzntxjd193ryZdpoWuYBXmgKD3KVw41uycsqAHH92VcU53rpDPJk")
	sig2, _ := solana.SignatureFromBase58("2EdJBKNNfodAsy6nFasWqYMtgFgU6rgY8WeckBMRHtgBEfdsn2pnKJfcj4LKABf76UonsTtn8UsPn7ZcbXMXkB5U")

	fromNativeAccount1, _ := solana.PublicKeyFromBase58("GmKsGRytiVoeMZGmBVCWPcUzJGHVqcvzhP5K9cstdr3E")
	fromNativeAccount2, _ := solana.PublicKeyFromBase58("5D4MWh35wxUcY1hBsm5GwuippPL2UBmfDnfkC8MeqxcN")
	tokenMintAccount, _ := solana.PublicKeyFromBase58(t.tokenConfig.Mint)
	toNativeAccount1, _ := solana.PublicKeyFromBase58("27htRMGeQ4HV32SPHsJrpndZn1zwmF2kiPqmABcHgehx")
	fromTokenAccount, _, _ := solana.FindAssociatedTokenAddress(fromNativeAccount1, tokenMintAccount)
	toTokenAccount1, _, _ := solana.FindAssociatedTokenAddress(toNativeAccount1, tokenMintAccount)

	computeBudgetPriceInst := computebudget.NewSetComputeUnitPriceInstructionBuilder().SetMicroLamports(1000).Build()
	transferCheckedInst1 := token.NewTransferCheckedInstructionBuilder().
		SetAmount(1).
		SetDecimals(uint8(t.tokenConfig.Decimals)).
		SetSourceAccount(fromTokenAccount).
		SetMintAccount(tokenMintAccount).
		SetDestinationAccount(toTokenAccount1).
		SetOwnerAccount(fromNativeAccount1).
		Build()

	transferInst := system.NewTransferInstructionBuilder().
		SetLamports(1).
		SetFundingAccount(fromNativeAccount2).
		SetRecipientAccount(fromNativeAccount1).
		Build()

	transferCheckedInsts = append(transferCheckedInsts, computeBudgetPriceInst)
	transferCheckedInsts = append(transferCheckedInsts, transferCheckedInst1)
	transferCheckedInsts = append(transferCheckedInsts, transferInst)
	transferCheckedTx, _ := solana.NewTransaction(transferCheckedInsts, refBlockHash, solana.TransactionPayer(fromNativeAccount1))
	transferCheckedTx.Signatures = append(transferCheckedTx.Signatures, sig1)
	transferCheckedTx.Signatures = append(transferCheckedTx.Signatures, sig2)

	accounts := []*solana.AccountMeta{solana.Meta(fromNativeAccount1).SIGNER().WRITE()}
	data := []byte("hello!!!")
	memoInst := solana.NewInstruction(solana.MemoProgramID, accounts, data)
	memoInsts = append(memoInsts, memoInst)
	memoInsts = append(memoInsts, transferInst)
	memoTx, _ := solana.NewTransaction(memoInsts, refBlockHash, solana.TransactionPayer(fromNativeAccount1))
	memoTx.Signatures = append(memoTx.Signatures, sig1)
	memoTx.Signatures = append(memoTx.Signatures, sig2)

	firstRun := true
	for {
		if firstRun {
			time.Sleep(10 * time.Second)
			firstRun = false
		} else {
			time.Sleep(10 * time.Second)
		}

		func() {
			// 获取分布式锁（防止重复评估计算单元）
			mutex := t.redSync.NewMutex("SOL-ESTIMATE-COMPUTE-UNIT")
			if err := mutex.Lock(); err != nil {
				var errTaken *redsync.ErrTaken
				if errors.As(err, &errTaken) { // 过滤"锁已被占用"的预期错误[1](@ref)
					return
				}
				log.Errorf("获取计算单元评估锁失败: %v", err)
				return
			}
			defer mutex.Unlock() // 确保锁释放

			miniRent, err := t.rpcClient.GetMinimumBalanceForRentExemption(
				context.Background(),
				tokenregistry.TOKEN_META_SIZE,
				rpc.CommitmentFinalized,
			)
			if err != nil {
				log.Errorf("solana周期性评估计算单元 获取账户最小租金错误: %v", err)
				return
			}

			if err := t.redis.Set(context.Background(), "COMPUTE-UNIT-ASSOCIATED-ACCOUNT-MINI-RENT", miniRent, 1*time.Hour).Err(); err != nil {
				log.Errorf("solana周期性评估 - 更新最小账户租金到 - 刷新Redis错误: %v", err)
				return
			}

			createAccountInstruction := system.NewCreateAccountInstruction(
				miniRent,
				tokenregistry.TOKEN_META_SIZE,
				tokenregistry.ProgramID(),
				fromNativeAccount2,
				solana.NewWallet().PublicKey(),
			).Build()
			var associatedAccount []solana.Instruction
			associatedAccount = append(associatedAccount, computeBudgetPriceInst)
			associatedAccount = append(associatedAccount, createAccountInstruction)
			associatedTx, _ := solana.NewTransaction(associatedAccount, refBlockHash, solana.TransactionPayer(fromNativeAccount2))
			associatedTx.Signatures = append(associatedTx.Signatures, sig1)
			associatedTx.Signatures = append(associatedTx.Signatures, sig2)

			srAssociateTokenAccount, err := t.rpcClient.SimulateTransactionWithOpts(context.Background(), associatedTx, &rpc.SimulateTransactionOpts{ReplaceRecentBlockhash: true})
			if err != nil {
				log.Errorf("solana周期性评估 associate token account 的Compute Unit错误: %v", err)
				return
			}
			if err := t.redis.Set(context.Background(), "COMPUTE-UNIT-ASSOCIATED-ACCOUNT", *srAssociateTokenAccount.Value.UnitsConsumed, 1*time.Hour).Err(); err != nil {
				log.Errorf("solana周期性更新优先费用，刷新Redis错误: %v", err)
				return
			}

			srTransferChecked, err := t.rpcClient.SimulateTransactionWithOpts(context.Background(), transferCheckedTx, &rpc.SimulateTransactionOpts{ReplaceRecentBlockhash: true})
			if err != nil {
				log.Errorf("solana周期性评估 compute unit 1 错误: %v", err)
				return
			}
			log.Infof("solana周期性更新优先费用 获取transfer checked unit consumed %d", *srTransferChecked.Value.UnitsConsumed)
			if err := t.redis.Set(context.Background(), "COMPUTE-UNIT-TRANSFER-CHECKED-1", *srTransferChecked.Value.UnitsConsumed, 1*time.Hour).Err(); err != nil {
				log.Errorf("solana周期性更新优先费用，刷新Redis错误: %v", err)
				return
			}

			srMemo, err := t.rpcClient.SimulateTransactionWithOpts(context.Background(), memoTx, &rpc.SimulateTransactionOpts{ReplaceRecentBlockhash: true})
			if err != nil {
				log.Errorf("solana周期性评估 memo 错误: %v", err)
				return
			}
			if err := t.redis.Set(context.Background(), "COMPUTE-UNIT-MEMO", *srMemo.Value.UnitsConsumed, 1*time.Hour).Err(); err != nil {
				log.Errorf("solana周期性更新优先费用，刷新Redis错误: %v", err)
				return
			}
		}()
	}
}
