package stake

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"oshit-go/app/pos/api/types"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// AnchorDiscriminator 计算 Anchor global 指令判别器（SHA256("global:<name>") 前 8 字节）
func AnchorDiscriminator(name string) []byte {
	h := sha256.Sum256([]byte("global:" + name))
	return h[:8]
}

// StakeTxParser 质押交易解析器
type StakeTxParser struct {
	rpcClient *rpc.Client
	tokenDec  float64
	programId string
}

// NewStakeTxParser 创建新的解析器
func NewStakeTxParser(rpcClient *rpc.Client, tokenDec float64, programId string) *StakeTxParser {
	return &StakeTxParser{
		rpcClient: rpcClient,
		tokenDec:  tokenDec,
		programId: programId,
	}
}

// ParseStakeTx 解析质押交易
func (p *StakeTxParser) ParseStakeTx(ctx context.Context, txSignature solana.Signature) (*types.ParsedStakeTx, error) {
	var maxSupportVersion uint64 = 0

	tr, err := p.rpcClient.GetTransaction(
		ctx,
		txSignature,
		&rpc.GetTransactionOpts{
			MaxSupportedTransactionVersion: &maxSupportVersion,
			Commitment:                     rpc.CommitmentFinalized,
			Encoding:                       solana.EncodingBase64,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if tr == nil {
		return nil, fmt.Errorf("transaction not found")
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(tr.Transaction.GetBinary()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}

	instructions := tx.Message.Instructions
	if len(instructions) == 0 {
		return nil, fmt.Errorf("no instructions found in transaction")
	}

	var stakeInst solana.CompiledInstruction
	stakeInstExist := false
	for index, inst := range instructions {
		if int(inst.ProgramIDIndex) > len(tx.Message.AccountKeys)-1 {
			continue
		}
		programId := tx.Message.AccountKeys[inst.ProgramIDIndex]
		if programId.String() == p.programId {
			stakeInstExist = true
			stakeInst = instructions[index]
		}
	}
	if !stakeInstExist {
		return nil, fmt.Errorf("can not find stake inst in transaction")
	}
	instData := stakeInst.Data

	// 解析指令数据
	instructionData, err := p.parseInstructionData(instData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse instruction data: %w", err)
	}

	// 解析账户信息
	accounts, err := p.parseAccounts(stakeInst, tx.Message.AccountKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to parse accounts: %w", err)
	}

	// 构建结果
	result := &types.ParsedStakeTx{
		TransactionSignature: txSignature,
		InstructionData:      *instructionData,
		Accounts:             *accounts,
		Signatures:           tx.Signatures,
		Success:              tr.Meta.Err == nil,
		RawData:              instData,
	}

	return result, nil
}

// parseInstructionData 解析指令数据
// stake:   disc(8) + amount(8) + stakeType(1) = 17 bytes
// unstake: disc(8) + stakeIndex(1)            =  9 bytes
// restake: disc(8) + stakeIndex(1) + stakeType(1) = 10 bytes
func (p *StakeTxParser) parseInstructionData(data []byte) (*types.StakeInstructionData, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("instruction data too short: got %d bytes, expected at least 8", len(data))
	}

	result := &types.StakeInstructionData{
		Discriminator: data[:8],
	}
	// stake 指令携带 amount 和 stakeType
	if len(data) == 17 {
		result.Amount = binary.LittleEndian.Uint64(data[8:16])
		result.StakeType = data[16]
	}
	return result, nil
}

// parseAccounts 解析账户信息
// 各指令账户布局：
//
//	stake/unstake (≥7): [0]Staker [1]Deployer [2]Config [3]StakeInfo [4]StakeAccount [5]UserToken [6]Mint ...
//	restake       (6):  [0]Staker [1]Deployer [2]Config [3]StakeInfo [4]Mint [5]SystemProgram
func (p *StakeTxParser) parseAccounts(inst solana.CompiledInstruction, accountKeys []solana.PublicKey) (*types.StakeAccounts, error) {
	n := len(inst.Accounts)
	if n < 4 {
		return nil, fmt.Errorf("insufficient accounts in instruction: got %d, expected at least 4", n)
	}

	get := func(i int) solana.PublicKey {
		if i < n {
			return accountKeys[inst.Accounts[i]]
		}
		return solana.PublicKey{}
	}

	accounts := &types.StakeAccounts{
		Staker:           get(0),
		Deployer:         get(1),
		ConfigAccount:    get(2),
		StakeInfoAccount: get(3),
		ProgramID:        accountKeys[inst.ProgramIDIndex],
	}

	if n >= 7 {
		// stake / unstake：StakeAccount(4) UserToken(5) Mint(6)
		accounts.StakeAccount = get(4)
		accounts.UserTokenAccount = get(5)
		accounts.MintAccount = get(6)
	} else if n >= 5 {
		// restake：Mint(4)
		accounts.MintAccount = get(4)
	}

	return accounts, nil
}

// ValidateDiscriminator 验证discriminator是否匹配
func (p *StakeTxParser) ValidateDiscriminator(parsedTx *types.ParsedStakeTx, expectedDiscriminator []byte) bool {
	if len(parsedTx.InstructionData.Discriminator) != 8 || len(expectedDiscriminator) != 8 {
		return false
	}

	for i := 0; i < 8; i++ {
		if parsedTx.InstructionData.Discriminator[i] != expectedDiscriminator[i] {
			return false
		}
	}
	return true
}

// GetStakeAmount 获取质押数量
func (p *StakeTxParser) GetStakeAmount(parsedTx *types.ParsedStakeTx) uint64 {
	return uint64(float64(parsedTx.InstructionData.Amount) * p.tokenDec)
}

// GetStakeType 获取质押类型
func (p *StakeTxParser) GetStakeType(parsedTx *types.ParsedStakeTx) uint8 {
	return parsedTx.InstructionData.StakeType
}

// IsTransactionSuccessful 检查交易是否成功
func (p *StakeTxParser) IsTransactionSuccessful(parsedTx *types.ParsedStakeTx) bool {
	return parsedTx.Success
}

// GetStakerAddress 获取质押者地址
func (p *StakeTxParser) GetStakerAddress(parsedTx *types.ParsedStakeTx) solana.PublicKey {
	return parsedTx.Accounts.Staker
}

// BatchParseStakeTx 批量解析质押交易
func (p *StakeTxParser) BatchParseStakeTx(ctx context.Context, txSignatures []solana.Signature) (map[solana.Signature]*types.ParsedStakeTx, []error) {
	results := make(map[solana.Signature]*types.ParsedStakeTx)
	var errors []error

	for _, txSig := range txSignatures {
		parsedTx, err := p.ParseStakeTx(ctx, txSig)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to parse transaction %s: %w", txSig.String(), err))
			continue
		}
		results[txSig] = parsedTx
	}

	return results, errors
}
