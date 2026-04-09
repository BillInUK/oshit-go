package stake

import (
	"context"
	"encoding/binary"
	"fmt"
	"oshit-go/app/pos/api/types"

	"github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

//// StakeInstructionData 质押指令数据结构
//type StakeInstructionData struct {
//	Discriminator []byte // 8字节
//	Amount        uint64 // 质押数量
//	StakeType     uint8  // 质押类型
//}
//
//// StakeAccounts 质押指令相关的账户
//type StakeAccounts struct {
//	Staker           solana.PublicKey // 质押者
//	Deployer         solana.PublicKey // 部署者
//	ConfigAccount    solana.PublicKey // 配置账户
//	StakeInfoAccount solana.PublicKey // 质押信息账户
//	StakeAccount     solana.PublicKey // 质押账户
//	UserTokenAccount solana.PublicKey // 用户代币账户
//	MintAccount      solana.PublicKey // Mint账户
//	ProgramID        solana.PublicKey // 程序ID
//}

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
func (p *StakeTxParser) parseInstructionData(data []byte) (*types.StakeInstructionData, error) {
	if len(data) != 17 {
		return nil, fmt.Errorf("unexpected instruction data length: got %d, expected 17", len(data))
	}

	return &types.StakeInstructionData{
		Discriminator: data[:8],
		Amount:        binary.LittleEndian.Uint64(data[8:16]),
		StakeType:     data[16],
	}, nil
}

// parseAccounts 解析账户信息
func (p *StakeTxParser) parseAccounts(inst solana.CompiledInstruction, accountKeys []solana.PublicKey) (*types.StakeAccounts, error) {
	if len(inst.Accounts) < 7 {
		return nil, fmt.Errorf("insufficient accounts in instruction: got %d, expected at least 7", len(inst.Accounts))
	}

	return &types.StakeAccounts{
		Staker:           accountKeys[inst.Accounts[0]],
		Deployer:         accountKeys[inst.Accounts[1]],
		ConfigAccount:    accountKeys[inst.Accounts[2]],
		StakeInfoAccount: accountKeys[inst.Accounts[3]],
		StakeAccount:     accountKeys[inst.Accounts[4]],
		UserTokenAccount: accountKeys[inst.Accounts[5]],
		MintAccount:      accountKeys[inst.Accounts[6]],
		ProgramID:        accountKeys[inst.ProgramIDIndex],
	}, nil
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
