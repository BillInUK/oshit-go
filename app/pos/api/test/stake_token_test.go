package test

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
	"oshit-go/app/pos/api/types"
)

// ============================================================
// 质押合约常量
// ============================================================

const (
	StakingProgramIDStr = "CyLTEgvmqVF9dPJkT6bMgccfXL7G26EXRAM9FEuP5ki6"

	// devnet 测试环境下管理员与用户使用同一密钥
	StakingAdminPrivate = "5LCaLqUSWKaD95BsR6sNQ4A62rCoADsxVsfYqhirEsGa5eNmDZwx1vxWoDTgio8eKT3K9HHwj7b5KfaVjYEsin6a"

	// 每次质押的基础数量（不含精度，合约最小要求 100000）
	StakingBaseAmount uint64 = 100000

	// 质押类型 0 = 3分钟锁定（450 slots × 400ms ≈ 180s）
	StakingTypeShort uint8 = 0

	// 合约指令预留 CU（Anchor 程序 200000 通常足够）
	StakingCULimit uint32 = 200_000
)

// ============================================================
// 通用 Anchor 指令结构体（实现 solana.Instruction 接口）
// ============================================================

type stakingAnchorInst struct {
	programID solana.PublicKey
	accounts  []*solana.AccountMeta
	data      []byte
}

func (i *stakingAnchorInst) ProgramID() solana.PublicKey     { return i.programID }
func (i *stakingAnchorInst) Accounts() []*solana.AccountMeta { return i.accounts }
func (i *stakingAnchorInst) Data() ([]byte, error)           { return i.data, nil }

// ============================================================
// 辅助函数
// ============================================================

// stakingDisc 计算 Anchor global 指令判别器（SHA256("global:<name>") 前 8 字节）
func stakingDisc(name string) []byte {
	h := sha256.Sum256([]byte("global:" + name))
	return h[:8]
}

// stakingComputePDAs 计算质押合约相关 PDA 地址
func stakingComputePDAs(programID, userPubKey, mintPubKey solana.PublicKey) (
	configPDA, stakeInfoPDA, stakeAcctPDA, userTokenAcct solana.PublicKey,
) {
	configPDA, _, _ = solana.FindProgramAddress(
		[][]byte{[]byte("config")},
		programID,
	)
	stakeInfoPDA, _, _ = solana.FindProgramAddress(
		[][]byte{[]byte("stake_info"), userPubKey.Bytes()},
		programID,
	)
	stakeAcctPDA, _, _ = solana.FindProgramAddress(
		[][]byte{[]byte("token"), userPubKey.Bytes()},
		programID,
	)
	userTokenAcct, _, _ = solana.FindAssociatedTokenAddress(userPubKey, mintPubKey)
	return
}

// stakingBuildStakeInst 构建 stake 指令
//
// Accounts（按合约 Stake 结构体顺序）：
//
//	0  signer           writable, signer
//	1  admin_signer     signer
//	2  config           PDA, read-only
//	3  stake_info       PDA, writable (init_if_needed)
//	4  stake_account    PDA, writable (init_if_needed)
//	5  user_token_acct  ATA, writable
//	6  mint             read-only
//	7  token_program    read-only
//	8  associated_token read-only
//	9  system_program   read-only
//
// Data: disc(8) + amount:u64(LE) + stake_type:u8
func stakingBuildStakeInst(
	programID, userPubKey, adminPubKey,
	configPDA, stakeInfoPDA, stakeAcctPDA, userTokenAcct, mintPubKey solana.PublicKey,
	amount uint64, stakeType uint8,
) solana.Instruction {
	data := make([]byte, 8+8+1)
	copy(data[:8], stakingDisc("stake"))
	binary.LittleEndian.PutUint64(data[8:16], amount)
	data[16] = stakeType

	return &stakingAnchorInst{
		programID: programID,
		accounts: []*solana.AccountMeta{
			{PublicKey: userPubKey, IsWritable: true, IsSigner: true},
			{PublicKey: adminPubKey, IsWritable: false, IsSigner: true},
			{PublicKey: configPDA, IsWritable: false, IsSigner: false},
			{PublicKey: stakeInfoPDA, IsWritable: true, IsSigner: false},
			{PublicKey: stakeAcctPDA, IsWritable: true, IsSigner: false},
			{PublicKey: userTokenAcct, IsWritable: true, IsSigner: false},
			{PublicKey: mintPubKey, IsWritable: false, IsSigner: false},
			{PublicKey: solana.TokenProgramID, IsWritable: false, IsSigner: false},
			{PublicKey: solana.SPLAssociatedTokenAccountProgramID, IsWritable: false, IsSigner: false},
			{PublicKey: solana.SystemProgramID, IsWritable: false, IsSigner: false},
		},
		data: data,
	}
}

// stakingBuildUnstakeInst 构建 unstake 指令
//
// Accounts（按合约 DeStake 结构体顺序）：
//
//	0  signer           writable, signer
//	1  admin_signer     signer
//	2  config           PDA, read-only
//	3  stake_info       PDA, writable
//	4  stake_account    PDA, writable
//	5  user_token_acct  ATA, writable
//	6  mint             read-only
//	7  token_program    read-only
//	8  associated_token read-only
//	9  system_program   read-only
//
// Data: disc(8) + stake_index:u8
func stakingBuildUnstakeInst(
	programID, userPubKey, adminPubKey,
	configPDA, stakeInfoPDA, stakeAcctPDA, userTokenAcct, mintPubKey solana.PublicKey,
	stakeIndex uint8,
) solana.Instruction {
	data := make([]byte, 8+1)
	copy(data[:8], stakingDisc("unstake"))
	data[8] = stakeIndex

	return &stakingAnchorInst{
		programID: programID,
		accounts: []*solana.AccountMeta{
			{PublicKey: userPubKey, IsWritable: true, IsSigner: true},
			{PublicKey: adminPubKey, IsWritable: false, IsSigner: true},
			{PublicKey: configPDA, IsWritable: false, IsSigner: false},
			{PublicKey: stakeInfoPDA, IsWritable: true, IsSigner: false},
			{PublicKey: stakeAcctPDA, IsWritable: true, IsSigner: false},
			{PublicKey: userTokenAcct, IsWritable: true, IsSigner: false},
			{PublicKey: mintPubKey, IsWritable: false, IsSigner: false},
			{PublicKey: solana.TokenProgramID, IsWritable: false, IsSigner: false},
			{PublicKey: solana.SPLAssociatedTokenAccountProgramID, IsWritable: false, IsSigner: false},
			{PublicKey: solana.SystemProgramID, IsWritable: false, IsSigner: false},
		},
		data: data,
	}
}

// stakingBuildRestakeInst 构建 restake 指令
//
// Accounts（按合约 ReStake 结构体顺序）：
//
//	0  signer        writable, signer
//	1  admin_signer  signer
//	2  config        PDA, read-only
//	3  stake_info    PDA, writable
//	4  mint          read-only
//	5  system_prog   read-only
//
// Data: disc(8) + stake_index:u8 + stake_type:u8
func stakingBuildRestakeInst(
	programID, userPubKey, adminPubKey,
	configPDA, stakeInfoPDA, mintPubKey solana.PublicKey,
	stakeIndex, stakeType uint8,
) solana.Instruction {
	data := make([]byte, 8+1+1)
	copy(data[:8], stakingDisc("restake"))
	data[8] = stakeIndex
	data[9] = stakeType

	return &stakingAnchorInst{
		programID: programID,
		accounts: []*solana.AccountMeta{
			{PublicKey: userPubKey, IsWritable: true, IsSigner: true},
			{PublicKey: adminPubKey, IsWritable: false, IsSigner: true},
			{PublicKey: configPDA, IsWritable: false, IsSigner: false},
			{PublicKey: stakeInfoPDA, IsWritable: true, IsSigner: false},
			{PublicKey: mintPubKey, IsWritable: false, IsSigner: false},
			{PublicKey: solana.SystemProgramID, IsWritable: false, IsSigner: false},
		},
		data: data,
	}
}

// stakingSendAndConfirm 构建交易（含优先费）、多签、发送，并轮询链上确认。
// userKey 和 adminKey 可以是同一密钥（devnet 测试场景）。
func stakingSendAndConfirm(
	ctx context.Context,
	t *testing.T,
	label string,
	insts []solana.Instruction,
	userKey, adminKey solana.PrivateKey,
) solana.Signature {
	t.Helper()

	// 拉取优先费（失败时用默认值）
	computePrice, err := getPriorityFee()
	if err != nil {
		t.Logf("[%s] getPriorityFee 失败，使用默认值 1000: %v", label, err)
		computePrice = 1000
	}

	// 优先费指令放在最前面
	allInsts := []solana.Instruction{
		computebudget.NewSetComputeUnitPriceInstructionBuilder().SetMicroLamports(computePrice).Build(),
		computebudget.NewSetComputeUnitLimitInstructionBuilder().SetUnits(StakingCULimit).Build(),
	}
	allInsts = append(allInsts, insts...)

	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		t.Fatalf("[%s] GetLatestBlockhash 失败: %v", label, err)
	}

	payer := userKey.PublicKey()
	tx, err := solana.NewTransaction(allInsts, recent.Value.Blockhash, solana.TransactionPayer(payer))
	if err != nil {
		t.Fatalf("[%s] NewTransaction 失败: %v", label, err)
	}

	// 构建签名者 map（同一密钥只签一次）
	keyMap := map[string]solana.PrivateKey{
		userKey.PublicKey().String():  userKey,
		adminKey.PublicKey().String(): adminKey,
	}
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		sk, ok := keyMap[key.String()]
		if !ok {
			return nil
		}
		return &sk
	})
	if err != nil {
		t.Fatalf("[%s] 签名失败: %v", label, err)
	}

	sig, err := rpcClient.SendTransaction(ctx, tx)
	if err != nil {
		t.Fatalf("[%s] SendTransaction 失败: %v", label, err)
	}
	fmt.Printf("[%s] 已提交: https://explorer.solana.com/tx/%s?cluster=devnet\n", label, sig.String())

	// 轮询等待确认（最多 60 秒）
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		statuses, err := rpcClient.GetSignatureStatuses(ctx, false, sig)
		if err != nil || statuses == nil || len(statuses.Value) == 0 || statuses.Value[0] == nil {
			continue
		}
		st := statuses.Value[0]
		if st.Err != nil {
			t.Fatalf("[%s] 链上执行失败: %v", label, st.Err)
		}
		if st.ConfirmationStatus == rpc.ConfirmationStatusFinalized ||
			st.ConfirmationStatus == rpc.ConfirmationStatusConfirmed {
			fmt.Printf("[%s] 链上确认成功 (status=%s)\n", label, st.ConfirmationStatus)
			return sig
		}
	}
	t.Fatalf("[%s] 等待链上确认超时（60 秒）", label)
	return sig
}

// ============================================================
// 测试用例
// ============================================================

// TestStakeUnstakeRestake 测试质押合约的完整流程：
//
//  1. 连续 stake 3 次（类型 0，3分钟锁定）
//  2. 等待锁定期结束（~3分10秒）
//  3. unstake 索引 0（token 退回用户）
//  4. restake 索引 1（原地续期，token 不移动）
//  5. 索引 2 保持质押状态不变
//
// 注意：本测试总耗时约 4–5 分钟，请确保 go test -timeout 10m。
func TestStakeUnstakeRestake(t *testing.T) {
	ctx := context.Background()

	userKey, err := solana.PrivateKeyFromBase58(DavidPrivate)
	if err != nil {
		t.Fatalf("解析用户私钥失败: %v", err)
	}
	adminKey, err := solana.PrivateKeyFromBase58(StakingAdminPrivate)
	if err != nil {
		t.Fatalf("解析管理员私钥失败: %v", err)
	}

	programID := solana.MPK(StakingProgramIDStr)
	mintPubKey := solana.MPK(TokenMintAddress)
	userPubKey := userKey.PublicKey()
	adminPubKey := adminKey.PublicKey()

	configPDA, stakeInfoPDA, stakeAcctPDA, userTokenAcct :=
		stakingComputePDAs(programID, userPubKey, mintPubKey)

	fmt.Println("=== 质押合约测试 ===")
	fmt.Printf("用户地址      : %s\n", userPubKey.String())
	fmt.Printf("管理员地址    : %s\n", adminPubKey.String())
	fmt.Printf("Config PDA    : %s\n", configPDA.String())
	fmt.Printf("StakeInfo PDA : %s\n", stakeInfoPDA.String())
	fmt.Printf("StakeAcct PDA : %s\n", stakeAcctPDA.String())
	fmt.Printf("UserTokenAcct : %s\n", userTokenAcct.String())

	// --------------------------------------------------------
	// Phase 1: Stake × 3（类型 0，3分钟锁定期）
	// --------------------------------------------------------
	fmt.Println("\n=== Phase 1: Stake × 3 ===")

	for i := 0; i < 3; i++ {
		inst := stakingBuildStakeInst(
			programID, userPubKey, adminPubKey,
			configPDA, stakeInfoPDA, stakeAcctPDA, userTokenAcct, mintPubKey,
			StakingBaseAmount, StakingTypeShort,
		)
		label := fmt.Sprintf("Stake[%d]", i)
		stakingSendAndConfirm(ctx, t, label, []solana.Instruction{inst}, userKey, adminKey)
		fmt.Printf("[%s] 质押 %d token 成功，锁定到 ~3分钟后\n", label, StakingBaseAmount)

		// 稍作间隔避免 blockhash 复用失败
		time.Sleep(500 * time.Millisecond)
	}

	// --------------------------------------------------------
	// Phase 2: 等待锁定期结束
	// 450 slots × 400ms = 180s，保留 10s 余量
	// --------------------------------------------------------
	lockWait := 190 * time.Second
	fmt.Printf("\n=== Phase 2: 等待锁定期结束（%v）===\n", lockWait)

	deadline := time.Now().Add(lockWait)
	for time.Now().Before(deadline) {
		remaining := time.Until(deadline).Round(time.Second)
		fmt.Printf("  剩余: %v ...\n", remaining)
		if remaining < 30*time.Second {
			time.Sleep(remaining)
		} else {
			time.Sleep(30 * time.Second)
		}
	}
	fmt.Println("  锁定期已结束，开始后续操作")

	// --------------------------------------------------------
	// Phase 3: Unstake 索引 0 —— token 退回用户账户
	// --------------------------------------------------------
	fmt.Println("\n=== Phase 3: Unstake[0] ===")

	unstakeInst := stakingBuildUnstakeInst(
		programID, userPubKey, adminPubKey,
		configPDA, stakeInfoPDA, stakeAcctPDA, userTokenAcct, mintPubKey,
		0,
	)
	stakingSendAndConfirm(ctx, t, "Unstake[0]", []solana.Instruction{unstakeInst}, userKey, adminKey)
	fmt.Println("  索引 0 解除质押成功，token 已退回用户")

	// --------------------------------------------------------
	// Phase 4: Restake 索引 1 —— 原地续期，token 不移动
	// --------------------------------------------------------
	fmt.Println("\n=== Phase 4: Restake[1] ===")

	restakeInst := stakingBuildRestakeInst(
		programID, userPubKey, adminPubKey,
		configPDA, stakeInfoPDA, mintPubKey,
		1, StakingTypeShort,
	)
	stakingSendAndConfirm(ctx, t, "Restake[1]", []solana.Instruction{restakeInst}, userKey, adminKey)
	fmt.Println("  索引 1 重新质押成功，锁定期已重置（token 原地不动）")

	// --------------------------------------------------------
	// 结果汇总
	// --------------------------------------------------------
	fmt.Println("\n=== 测试完成 ===")
	fmt.Printf("  索引 0: unstake —— staked_amount 已清零，token 退回用户\n")
	fmt.Printf("  索引 1: restake —— 锁定期已续期，token 仍在合约 PDA 中\n")
	fmt.Printf("  索引 2: 保持质押状态，锁定期已过（待后续手动操作）\n")
}

// ============================================================
// 链上数据解析辅助
// ============================================================

// stakeRecordBorshSize 是单条 StakeRecord 的 borsh 序列化字节数（无 padding）：
// stake_type(1) + staked_amount(8) + stake_start_slot(8) + stake_end_slot(8) = 25
const stakeRecordBorshSize = 25

// parsedStakeRecord 解析后的链上质押记录
type parsedStakeRecord struct {
	Index          uint8
	StakeType      uint8
	StakedAmount   uint64
	StakeStartSlot uint64
	StakeEndSlot   uint64
}

// stakingReadStakeInfo 从链上读取并解析 stake_info PDA 账户数据。
//
// 账户数据布局（borsh）：
//
//	[0:8]         discriminator
//	[8:40]        user_wallet (Pubkey, 32 bytes)
//	[40:40+N*25]  stakes[N]   (每条 25 bytes, N=MAX_STAKE_RECORDS=10)
func stakingReadStakeInfo(ctx context.Context, stakeInfoPDA solana.PublicKey) ([]parsedStakeRecord, error) {
	info, err := rpcClient.GetAccountInfo(ctx, stakeInfoPDA)
	if err != nil {
		return nil, fmt.Errorf("GetAccountInfo error: %w", err)
	}
	if info == nil || info.Value == nil {
		return nil, fmt.Errorf("stake_info 账户不存在（用户尚未质押过）")
	}
	data := info.Value.Data.GetBinary()

	const maxRecords = 10
	minLen := 8 + 32 + maxRecords*stakeRecordBorshSize // = 290
	if len(data) < minLen {
		return nil, fmt.Errorf("账户数据长度不足: %d < %d", len(data), minLen)
	}

	records := make([]parsedStakeRecord, maxRecords)
	offset := 40 // 跳过 discriminator(8) + user_wallet(32)
	for i := 0; i < maxRecords; i++ {
		records[i].Index = uint8(i)
		records[i].StakeType = data[offset]
		records[i].StakedAmount = binary.LittleEndian.Uint64(data[offset+1 : offset+9])
		records[i].StakeStartSlot = binary.LittleEndian.Uint64(data[offset+9 : offset+17])
		records[i].StakeEndSlot = binary.LittleEndian.Uint64(data[offset+17 : offset+25])
		offset += stakeRecordBorshSize
	}
	return records, nil
}

// ============================================================
// 后端接口辅助函数
// ============================================================

// stakingQueryAuthority 从链上 config PDA 读取合约管理员（authority）公钥。
//
// Config 账户数据布局（Anchor）：
//
//	[0:8]   discriminator
//	[8:40]  allowed_mint  (Pubkey, 32 bytes)
//	[40:72] authority     (Pubkey, 32 bytes)
//	[72:80] min_stake_amount (u64)
func stakingQueryAuthority(ctx context.Context, configPDA solana.PublicKey) (solana.PublicKey, error) {
	info, err := rpcClient.GetAccountInfo(ctx, configPDA)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("GetAccountInfo error: %w", err)
	}
	if info == nil || info.Value == nil {
		return solana.PublicKey{}, fmt.Errorf("config PDA 账户不存在，合约尚未初始化")
	}
	data := info.Value.Data.GetBinary()
	if len(data) < 72 {
		return solana.PublicKey{}, fmt.Errorf("config 账户数据长度不足: %d bytes", len(data))
	}
	var authority solana.PublicKey
	copy(authority[:], data[40:72])
	return authority, nil
}

// stakingBuildHexEncodedTx 构建质押相关交易并以 hex 返回。
//
// 测试侧只持有用户私钥，因此只签 Signatures[0]（用户位）。
// Signatures[1]（admin 位）留零，由 base 服务通过 t_service_key 中配置的
// 合约 authority 私钥补签（需要 t_service_info.multi_sign=true）。
//
// 返回：(hex 编码交易, txId=Signatures[0]字符串, error)
func stakingBuildHexEncodedTx(
	ctx context.Context,
	userKey solana.PrivateKey,
	insts []solana.Instruction,
) (encodedHex, txId string, err error) {
	computePrice, feeErr := getPriorityFee()
	if feeErr != nil || computePrice == 0 {
		computePrice = 1000
	}

	allInsts := []solana.Instruction{
		computebudget.NewSetComputeUnitPriceInstructionBuilder().SetMicroLamports(computePrice).Build(),
		computebudget.NewSetComputeUnitLimitInstructionBuilder().SetUnits(StakingCULimit).Build(),
	}
	allInsts = append(allInsts, insts...)

	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return "", "", fmt.Errorf("GetLatestBlockhash error: %w", err)
	}

	payer := userKey.PublicKey()
	tx, err := solana.NewTransaction(allInsts, recent.Value.Blockhash, solana.TransactionPayer(payer))
	if err != nil {
		return "", "", fmt.Errorf("NewTransaction error: %w", err)
	}

	// 按 numRequiredSignatures 初始化签名槽（全零）。
	// index 0 = 用户签名（测试侧可签）
	// index 1 = admin 签名（留零，由 base 服务用 service key 补签）
	numSigs := int(tx.Message.Header.NumRequiredSignatures)
	tx.Signatures = make([]solana.Signature, numSigs)

	msgBin, err := tx.Message.MarshalBinary()
	if err != nil {
		return "", "", fmt.Errorf("MarshalBinary error: %w", err)
	}
	userSig, err := userKey.Sign(msgBin)
	if err != nil {
		return "", "", fmt.Errorf("Sign error: %w", err)
	}
	tx.Signatures[0] = userSig
	txId = userSig.String()

	// 序列化为二进制并 hex 编码
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		return "", "", fmt.Errorf("tx.MarshalBinary error: %w", err)
	}
	return hex.EncodeToString(txBytes), txId, nil
}

// ============================================================
// 测试用例：通过后端接口提交 stake 交易
// ============================================================

// TestSubmitStakeTokenViaBackend 通过 StakeHandler.StakeToken 接口提交质押交易。
//
// 流程：
//  1. 查询链上 config PDA 获取合约 authority（admin 公钥）
//  2. 构建 stake 指令（含 authority 作为 admin_signer）
//  3. 仅用用户私钥签 Signatures[0]，留零给 base 服务补签
//  4. Hex 编码后 POST 到 /snap/stake/token/stake
//  5. 轮询链上确认（最多 60 秒）
func TestSubmitStakeTokenViaBackend(t *testing.T) {
	ctx := context.Background()

	userKey, err := solana.PrivateKeyFromBase58(DavidPrivate)
	if err != nil {
		t.Fatalf("解析用户私钥失败: %v", err)
	}
	userPubKey := userKey.PublicKey()

	programID := solana.MPK(StakingProgramIDStr)
	mintPubKey := solana.MPK(TokenMintAddress)

	// 1. 计算 PDAs
	configPDA, stakeInfoPDA, stakeAcctPDA, userTokenAcct :=
		stakingComputePDAs(programID, userPubKey, mintPubKey)

	fmt.Println("=== 后端接口质押测试 ===")
	fmt.Printf("用户地址      : %s\n", userPubKey.String())
	fmt.Printf("Config PDA    : %s\n", configPDA.String())
	fmt.Printf("StakeInfo PDA : %s\n", stakeInfoPDA.String())
	fmt.Printf("StakeAcct PDA : %s\n", stakeAcctPDA.String())
	fmt.Printf("UserTokenAcct : %s\n", userTokenAcct.String())

	// 2. 从链上 config PDA 获取 authority 作为 adminPubKey
	adminPubKey, err := stakingQueryAuthority(ctx, configPDA)
	if err != nil {
		t.Fatalf("查询合约 authority 失败: %v\n(请确认合约已在 devnet 上初始化)", err)
	}
	fmt.Printf("合约 authority: %s\n", adminPubKey.String())

	// 3. 构建 stake 指令
	stakeInst := stakingBuildStakeInst(
		programID, userPubKey, adminPubKey,
		configPDA, stakeInfoPDA, stakeAcctPDA, userTokenAcct, mintPubKey,
		StakingBaseAmount, StakingTypeShort,
	)

	// 4. 构建交易：用户签 Signatures[0]，其余签名槽留零
	encodedHex, txId, err := stakingBuildHexEncodedTx(ctx, userKey, []solana.Instruction{stakeInst})
	if err != nil {
		t.Fatalf("构建交易失败: %v", err)
	}
	fmt.Printf("TxID (用户签名) : %s\n", txId)
	fmt.Printf("EncodedTx 长度  : %d chars\n", len(encodedHex))

	// 5. 提交到后端接口
	rsp, err := postJsonRequest[any](
		SnapURL+"/stake/token/stake",
		types.EncodedTxReq{EncodedTx: encodedHex},
		nil,
	)
	if err != nil {
		t.Fatalf("请求后端接口失败: %v", err)
	}
	if rsp.Code != 200 {
		t.Fatalf("后端返回错误: code=%d msg=%s", rsp.Code, rsp.Msg)
	}
	fmt.Printf("后端接受成功 (code=%d)，等待链上确认...\n", rsp.Code)
	fmt.Printf("Solscan: https://explorer.solana.com/tx/%s?cluster=devnet\n", txId)

	// 6. 轮询链上确认（最多 60 秒）
	userSig, err := solana.SignatureFromBase58(txId)
	if err != nil {
		t.Fatalf("解析 txId 签名失败: %v", err)
	}
	for i := 0; i < 30; i++ {
		time.Sleep(2 * time.Second)
		statuses, err := rpcClient.GetSignatureStatuses(ctx, false, userSig)
		if err != nil || statuses == nil || len(statuses.Value) == 0 || statuses.Value[0] == nil {
			continue
		}
		st := statuses.Value[0]
		if st.Err != nil {
			t.Fatalf("链上执行失败: %v", st.Err)
		}
		if st.ConfirmationStatus == rpc.ConfirmationStatusFinalized ||
			st.ConfirmationStatus == rpc.ConfirmationStatusConfirmed {
			fmt.Printf("链上确认成功 (status=%s)\n", st.ConfirmationStatus)
			fmt.Printf("=== 质押完成，amount=%d token ===\n", StakingBaseAmount)
			return
		}
	}
	t.Fatalf("等待链上确认超时（60 秒），txId=%s", txId)
}

// ============================================================
// 测试用例：取出全部已到期的质押
// ============================================================

// TestUnstakeAll 读取链上所有质押记录，将全部已到期（stake_end_slot <= currentSlot）
// 且有余额（staked_amount > 0）的记录逐一解押，token 退回用户账户。
//
// 运行命令：
//
//	go test ./test/... -v -run TestUnstakeAll -timeout 5m
func TestUnstakeAll(t *testing.T) {
	ctx := context.Background()

	userKey, err := solana.PrivateKeyFromBase58(DavidPrivate)
	if err != nil {
		t.Fatalf("解析私钥失败: %v", err)
	}
	userPubKey := userKey.PublicKey()

	programID := solana.MPK(StakingProgramIDStr)
	mintPubKey := solana.MPK(TokenMintAddress)

	configPDA, stakeInfoPDA, stakeAcctPDA, userTokenAcct :=
		stakingComputePDAs(programID, userPubKey, mintPubKey)

	// 从链上 config PDA 读取合约 authority（即 admin 公钥）
	adminPubKey, err := stakingQueryAuthority(ctx, configPDA)
	if err != nil {
		t.Fatalf("查询合约 authority 失败: %v", err)
	}

	fmt.Printf("用户地址      : %s\n", userPubKey.String())
	fmt.Printf("合约 authority: %s\n", adminPubKey.String())
	fmt.Printf("StakeInfo PDA : %s\n", stakeInfoPDA.String())

	// 1. 读取链上 stake_info 账户并解析所有记录
	records, err := stakingReadStakeInfo(ctx, stakeInfoPDA)
	if err != nil {
		t.Fatalf("读取 stake_info 失败: %v", err)
	}

	// 2. 获取当前 slot
	currentSlot, err := rpcClient.GetSlot(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		t.Fatalf("GetSlot 失败: %v", err)
	}

	// 3. 打印所有记录并筛选可 unstake 的
	fmt.Printf("\n%-6s %-6s %-12s %-14s %-14s %s\n",
		"Index", "Type", "Amount", "StartSlot", "EndSlot", "Status")
	fmt.Println("--------------------------------------------------------------")

	var toUnstake []parsedStakeRecord
	for _, r := range records {
		if r.StakedAmount == 0 {
			fmt.Printf("%-6d %-6s %-12s %-14s %-14s %s\n",
				r.Index, "-", "-", "-", "-", "empty")
			continue
		}
		var status string
		if currentSlot >= r.StakeEndSlot {
			status = "expired ✓"
			toUnstake = append(toUnstake, r)
		} else {
			remainSec := (r.StakeEndSlot - currentSlot) * 400 / 1000
			status = fmt.Sprintf("locked (~%ds left)", remainSec)
		}
		fmt.Printf("%-6d %-6d %-12d %-14d %-14d %s\n",
			r.Index, r.StakeType, r.StakedAmount, r.StakeStartSlot, r.StakeEndSlot, status)
	}
	fmt.Printf("当前 slot: %d\n", currentSlot)

	if len(toUnstake) == 0 {
		t.Log("\n没有已到期的质押记录，跳过 unstake。")
		return
	}

	fmt.Printf("\n=== 开始 Unstake（共 %d 条到期记录）===\n", len(toUnstake))

	// 4. 逐一通过后端接口提交 unstake 交易
	for _, r := range toUnstake {
		label := fmt.Sprintf("Unstake[%d]", r.Index)

		inst := stakingBuildUnstakeInst(
			programID, userPubKey, adminPubKey,
			configPDA, stakeInfoPDA, stakeAcctPDA, userTokenAcct, mintPubKey,
			r.Index,
		)

		// 构建交易：用户签 Signatures[0]，hex 编码后发往后端
		encodedHex, txId, err := stakingBuildHexEncodedTx(ctx, userKey, []solana.Instruction{inst})
		if err != nil {
			t.Fatalf("[%s] 构建交易失败: %v", label, err)
		}
		fmt.Printf("[%s] TxID=%s\n", label, txId)

		// 提交到后端 /snap/stake/token/unstake
		rsp, err := postJsonRequest[any](
			SnapURL+"/stake/token/unstake",
			types.EncodedTxReq{EncodedTx: encodedHex},
			nil,
		)
		if err != nil {
			t.Fatalf("[%s] 请求后端接口失败: %v", label, err)
		}
		if rsp.Code != 200 {
			t.Fatalf("[%s] 后端返回错误: code=%d msg=%s", label, rsp.Code, rsp.Msg)
		}
		fmt.Printf("[%s] 后端接受成功，等待链上确认... https://explorer.solana.com/tx/%s?cluster=devnet\n", label, txId)

		// 轮询链上确认（最多 60 秒）
		userSig, err := solana.SignatureFromBase58(txId)
		if err != nil {
			t.Fatalf("[%s] 解析 txId 签名失败: %v", label, err)
		}
		confirmed := false
		for i := 0; i < 30; i++ {
			time.Sleep(2 * time.Second)
			statuses, err := rpcClient.GetSignatureStatuses(ctx, false, userSig)
			if err != nil || statuses == nil || len(statuses.Value) == 0 || statuses.Value[0] == nil {
				continue
			}
			st := statuses.Value[0]
			if st.Err != nil {
				t.Fatalf("[%s] 链上执行失败: %v", label, st.Err)
			}
			if st.ConfirmationStatus == rpc.ConfirmationStatusFinalized ||
				st.ConfirmationStatus == rpc.ConfirmationStatusConfirmed {
				fmt.Printf("[%s] 链上确认成功 (status=%s)，取回 %d token（type=%d）\n",
					label, st.ConfirmationStatus, r.StakedAmount, r.StakeType)
				confirmed = true
				break
			}
		}
		if !confirmed {
			t.Fatalf("[%s] 等待链上确认超时（60 秒），txId=%s", label, txId)
		}
	}

	fmt.Printf("\n=== 全部到期质押已取出（%d 条）===\n", len(toUnstake))
}

// ============================================================
// 测试用例：将全部已到期的质押通过后端接口重新质押
// ============================================================

// TestRestakeAll 读取链上所有质押记录，将全部已到期（stake_end_slot <= currentSlot）
// 且有余额（staked_amount > 0）的记录逐一重新质押（原地续期，token 不移动）。
// 交易通过后端接口转发，base 服务负责补签 admin 签名并广播。
//
// 运行命令：
//
//	go test ./test/... -v -run TestRestakeAll -timeout 5m
func TestRestakeAll(t *testing.T) {
	ctx := context.Background()

	userKey, err := solana.PrivateKeyFromBase58(DavidPrivate)
	if err != nil {
		t.Fatalf("解析私钥失败: %v", err)
	}
	userPubKey := userKey.PublicKey()

	programID := solana.MPK(StakingProgramIDStr)
	mintPubKey := solana.MPK(TokenMintAddress)

	configPDA, stakeInfoPDA, _, _ :=
		stakingComputePDAs(programID, userPubKey, mintPubKey)

	// 从链上 config PDA 读取合约 authority（即 admin 公钥）
	adminPubKey, err := stakingQueryAuthority(ctx, configPDA)
	if err != nil {
		t.Fatalf("查询合约 authority 失败: %v", err)
	}

	fmt.Printf("用户地址      : %s\n", userPubKey.String())
	fmt.Printf("合约 authority: %s\n", adminPubKey.String())
	fmt.Printf("StakeInfo PDA : %s\n", stakeInfoPDA.String())

	// 1. 读取链上 stake_info 账户并解析所有记录
	records, err := stakingReadStakeInfo(ctx, stakeInfoPDA)
	if err != nil {
		t.Fatalf("读取 stake_info 失败: %v", err)
	}

	// 2. 获取当前 slot
	currentSlot, err := rpcClient.GetSlot(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		t.Fatalf("GetSlot 失败: %v", err)
	}

	// 3. 打印所有记录并筛选可 restake 的（已到期且有余额）
	fmt.Printf("\n%-6s %-6s %-12s %-14s %-14s %s\n",
		"Index", "Type", "Amount", "StartSlot", "EndSlot", "Status")
	fmt.Println("--------------------------------------------------------------")

	var toRestake []parsedStakeRecord
	for _, r := range records {
		if r.StakedAmount == 0 {
			fmt.Printf("%-6d %-6s %-12s %-14s %-14s %s\n",
				r.Index, "-", "-", "-", "-", "empty")
			continue
		}
		var status string
		if currentSlot >= r.StakeEndSlot {
			status = "expired ✓"
			toRestake = append(toRestake, r)
		} else {
			remainSec := (r.StakeEndSlot - currentSlot) * 400 / 1000
			status = fmt.Sprintf("locked (~%ds left)", remainSec)
		}
		fmt.Printf("%-6d %-6d %-12d %-14d %-14d %s\n",
			r.Index, r.StakeType, r.StakedAmount, r.StakeStartSlot, r.StakeEndSlot, status)
	}
	fmt.Printf("当前 slot: %d\n", currentSlot)

	if len(toRestake) == 0 {
		t.Log("\n没有已到期的质押记录，跳过 restake。")
		return
	}

	fmt.Printf("\n=== 开始 Restake（共 %d 条到期记录）===\n", len(toRestake))

	// 4. 逐一通过后端接口提交 restake 交易
	for _, r := range toRestake {
		label := fmt.Sprintf("Restake[%d]", r.Index)

		inst := stakingBuildRestakeInst(
			programID, userPubKey, adminPubKey,
			configPDA, stakeInfoPDA, mintPubKey,
			r.Index, r.StakeType,
		)

		// 构建交易：用户签 Signatures[0]，admin 签名槽留零由后端补签
		encodedHex, txId, err := stakingBuildHexEncodedTx(ctx, userKey, []solana.Instruction{inst})
		if err != nil {
			t.Fatalf("[%s] 构建交易失败: %v", label, err)
		}
		fmt.Printf("[%s] TxID=%s\n", label, txId)

		// 提交到后端 /snap/stake/token/restake
		rsp, err := postJsonRequest[any](
			SnapURL+"/stake/token/restake",
			types.EncodedTxReq{EncodedTx: encodedHex},
			nil,
		)
		if err != nil {
			t.Fatalf("[%s] 请求后端接口失败: %v", label, err)
		}
		if rsp.Code != 200 {
			t.Fatalf("[%s] 后端返回错误: code=%d msg=%s", label, rsp.Code, rsp.Msg)
		}
		fmt.Printf("[%s] 后端接受成功，等待链上确认... https://explorer.solana.com/tx/%s?cluster=devnet\n", label, txId)

		// 轮询链上确认（最多 60 秒）
		userSig, err := solana.SignatureFromBase58(txId)
		if err != nil {
			t.Fatalf("[%s] 解析 txId 签名失败: %v", label, err)
		}
		confirmed := false
		for i := 0; i < 30; i++ {
			time.Sleep(2 * time.Second)
			statuses, err := rpcClient.GetSignatureStatuses(ctx, false, userSig)
			if err != nil || statuses == nil || len(statuses.Value) == 0 || statuses.Value[0] == nil {
				continue
			}
			st := statuses.Value[0]
			if st.Err != nil {
				t.Fatalf("[%s] 链上执行失败: %v", label, st.Err)
			}
			if st.ConfirmationStatus == rpc.ConfirmationStatusFinalized ||
				st.ConfirmationStatus == rpc.ConfirmationStatusConfirmed {
				fmt.Printf("[%s] 链上确认成功 (status=%s)，续期 %d token（type=%d）\n",
					label, st.ConfirmationStatus, r.StakedAmount, r.StakeType)
				confirmed = true
				break
			}
		}
		if !confirmed {
			t.Fatalf("[%s] 等待链上确认超时（60 秒），txId=%s", label, txId)
		}
	}

	fmt.Printf("\n=== 全部到期质押已续期（%d 条）===\n", len(toRestake))
}
