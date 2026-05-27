package utils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/mr-tron/base58"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"oshit-go/common/pkg/entity"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	SPLTokenProgramId       = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	DefaultComputeUnitLimit = 200_000
)

type SPLTokenAccountInfo struct {
	IsNative    bool             `json:"isNative"`
	Mint        solana.PublicKey `json:"mint"`
	Owner       solana.PublicKey `json:"owner"`
	State       string           `json:"state"`
	TokenAmount struct {
		Amount      string  `json:"amount"`
		Decimals    int     `json:"decimals"`
		UIAmount    float64 `json:"uiAmount"`
		UIAmountStr string  `json:"uiAmountString"`
	} `json:"tokenAmount"`
}

type SPLTokenAccountData struct {
	TokenAccountInfo SPLTokenAccountInfo `json:"info"`
	Type             string              `json:"type"`
}

type ParsedData struct {
	Parsed  SPLTokenAccountData `json:"parsed"`
	Program string              `json:"program"`
	Space   int                 `json:"space"`
}

// ── Helius Enhanced Transactions API ─────────────────────────────────────────

type heliusTokenBalanceChange struct {
	UserAccount    string `json:"userAccount"`
	TokenAccount   string `json:"tokenAccount"`
	Mint           string `json:"mint"`
	RawTokenAmount struct {
		TokenAmount string `json:"tokenAmount"`
		Decimals    uint8  `json:"decimals"`
	} `json:"rawTokenAmount"`
}

type heliusAccountData struct {
	Account             string                     `json:"account"`
	TokenBalanceChanges []heliusTokenBalanceChange `json:"tokenBalanceChanges"`
}

type heliusEnhancedTx struct {
	Signature        string              `json:"signature"`
	FeePayer         string              `json:"feePayer"`
	AccountData      []heliusAccountData `json:"accountData"`
	TransactionError interface{}         `json:"transactionError"`
}

func ParseDbArray(input string) []string {
	// 如果输入为空或长度小于2（不能构成花括号），直接返回空数组
	if len(input) < 2 {
		return []string{}
	}

	// 去掉开头和结尾的花括号
	input = strings.TrimSpace(input)
	if input[0] == '{' && input[len(input)-1] == '}' {
		input = input[1 : len(input)-1]
	} else {
		// 如果没有花括号，返回空数组
		return []string{}
	}

	// 如果去掉花括号后是空字符串，返回空数组
	if input == "" {
		return []string{}
	}

	// 使用逗号分割字符串，并去掉可能的引号
	items := strings.Split(input, ",")
	for i := range items {
		items[i] = strings.TrimSpace(items[i])
		items[i] = strings.Trim(items[i], "\"")
	}

	return items
}

// FilterSOLErrorMessage 过滤错误信息并返回人性化提示
func FilterSOLErrorMessage(originalErr error) error {
	errMsg := originalErr.Error()
	// 检查是否包含"insufficient funds for rent"错误
	if strings.Contains(errMsg, "insufficient funds for rent") {
		return fmt.Errorf("you don't have enough SOL")
	}
	if strings.Contains(errMsg, "Transaction simulation failed") {
		return fmt.Errorf("transaction failed, please try with a different wallet or address")
	}
	return originalErr
}

// GetCurrentSlot 获取当前slot信息
func GetCurrentSlot(rpcClient *rpc.Client, commitment rpc.CommitmentType) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slot, err := rpcClient.GetSlot(ctx, commitment)
	if err != nil {
		return 0, err
	}
	return slot, nil
}

func GetNativeAccountInfo(rpcClient *rpc.Client, nativeAccount solana.PublicKey) (*rpc.GetAccountInfoResult, error) {
	out, err := rpcClient.GetAccountInfo(context.Background(), nativeAccount)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GetSPLTokenAccounts 获取钱包的信息，包括钱包所有的tokenAccount，以及tokenAccount的余额
func GetSPLTokenAccounts(rpcClient *rpc.Client, nativeAccount solana.PublicKey) ([]solana.PublicKey, error) {
	var offset uint64 = 0
	var length uint64 = 10
	var tokenAccounts []solana.PublicKey
	walletTokenAccount, err := rpcClient.GetTokenAccountsByOwner(context.Background(), nativeAccount,
		&rpc.GetTokenAccountsConfig{
			ProgramId: &token.ProgramID,
		},
		&rpc.GetTokenAccountsOpts{
			Encoding: solana.EncodingBase64Zstd,
			DataSlice: &rpc.DataSlice{
				Offset: &offset,
				Length: &length,
			},
		})
	if err != nil {
		return nil, err
	}
	if len(walletTokenAccount.Value) <= 0 {
		return nil, errors.New("can not find any token account")
	}
	for _, account := range walletTokenAccount.Value {
		tokenAccounts = append(tokenAccounts, account.Pubkey)
	}
	return tokenAccounts, nil
}

func GetSPLTokenAccountInfo(rpcClient *rpc.Client, tokenAccount string) (*SPLTokenAccountInfo, error) {
	pubKey, err := solana.PublicKeyFromBase58(tokenAccount)
	if err != nil {
		return nil, errors.New("malformed token account,must in base58 format")
	}
	out, err := rpcClient.GetAccountInfoWithOpts(
		context.TODO(),
		pubKey,
		&rpc.GetAccountInfoOpts{
			Encoding: solana.EncodingJSONParsed,
		},
	)
	if err != nil {
		return nil, err
	}
	var parsedData ParsedData
	dataBytes, err := out.Value.Data.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(dataBytes, &parsedData)
	if err != nil {
		return nil, err
	}
	return &parsedData.Parsed.TokenAccountInfo, nil
}

// GetSPLTokenAccountOwner 获取钱包的TokenAccount
func GetSPLTokenAccountOwner(rpcClient *rpc.Client, tokenAccount solana.PublicKey) (*solana.PublicKey, error) {
	out, err := rpcClient.GetAccountInfoWithOpts(
		context.TODO(),
		tokenAccount,
		&rpc.GetAccountInfoOpts{
			Encoding: solana.EncodingJSONParsed,
		},
	)
	if err != nil {
		return nil, err
	}
	var parsedData ParsedData
	dataBytes, _ := out.Value.Data.MarshalJSON()

	err = json.Unmarshal(dataBytes, &parsedData)
	if err != nil {
		return nil, err
	}
	return &parsedData.Parsed.TokenAccountInfo.Owner, nil
}

func GetNativeAccountByTokenAccount(rpcClient *rpc.Client, tokenAccountPubKey solana.PublicKey) (*solana.PublicKey, error) {
	// Fetch account info
	accountInfo, err := rpcClient.GetAccountInfo(context.Background(), tokenAccountPubKey)
	if err != nil {
		fmt.Printf("Error retrieving account info: %v\n", err)
		return nil, err
	}
	// Ensure that data is available
	if accountInfo.Value == nil || accountInfo.Value.Data.GetBinary() == nil {
		fmt.Println("Account data is not available")
		return nil, err
	}
	data := accountInfo.Value.Data.GetBinary()
	// Check if data length is sufficient to contain Mint and Owner public keys
	if len(data) < 64 {
		return nil, errors.New("account data is too short to contain required information")
	}
	// Owner public key starts at byte 32 and spans 32 bytes
	ownerPubKey := solana.PublicKeyFromBytes(data[32:64])
	return &ownerPubKey, nil
}

func GetSPLTokenAccountByNative(rpcClient *rpc.Client, nativePubKey solana.PublicKey, tokenMintAccount solana.PublicKey) (*solana.PublicKey, error) {
	out, err := rpcClient.GetTokenAccountsByOwner(
		context.Background(),
		nativePubKey,
		&rpc.GetTokenAccountsConfig{
			Mint: &tokenMintAccount,
		},
		&rpc.GetTokenAccountsOpts{
			Encoding: solana.EncodingBase64Zstd,
		},
	)

	if err != nil {
		return nil, err
	}
	if len(out.Value) <= 0 {
		return nil, nil
	}
	return &out.Value[0].Pubkey, nil
}

// GetTxSigsForAddressWithBeforeSign 获取 [walletAddress] 在交易Id [beforeSig]之前的最多[limit]笔交易
func GetTxSigsForAddressWithBeforeSign(rpcClient *rpc.Client, accountPubKey solana.PublicKey, beforeSig solana.Signature, limit int) ([]*rpc.TransactionSignature, error) {
	// 查找before signature之前的交易
	txSign, err := rpcClient.GetSignaturesForAddressWithOpts(context.Background(),
		accountPubKey,
		&rpc.GetSignaturesForAddressOpts{
			Limit:      &limit,
			Before:     beforeSig,
			Commitment: rpc.CommitmentFinalized,
		})
	if err != nil {
		return nil, err
	}
	return txSign, nil
}

// GetAccountRecentTokenTransfer 查看地址最近的某种token的交易
func GetAccountRecentTokenTransfer(rpcClient *rpc.Client, nativeAccount, tokenMintAccount solana.PublicKey) ([]*rpc.TransactionSignature, error) {
	var limit = 1000
	tokenAccount, _, _ := solana.FindAssociatedTokenAddress(nativeAccount, tokenMintAccount)
	tokenTransactions, err := rpcClient.GetSignaturesForAddressWithOpts(context.Background(), tokenAccount, &rpc.GetSignaturesForAddressOpts{
		Limit:      &limit,
		Commitment: rpc.CommitmentFinalized,
	})
	if err != nil {
		return nil, err
	}
	return tokenTransactions, nil
}

// GetTransactionCountForReward  查看地址的交易数是否超过count
func GetTransactionCountForReward(rpcClient *rpc.Client, nativeAccount solana.PublicKey) (int, error) {
	// 查看地址的交易是否超过3笔
	var limit = 1000
	nativeTransactions, err := rpcClient.GetSignaturesForAddressWithOpts(context.Background(), nativeAccount, &rpc.GetSignaturesForAddressOpts{
		Limit:      &limit,
		Commitment: rpc.CommitmentFinalized,
	})
	if err != nil {
		return 0, err
	}
	// 获取所有的token account，然后获取每个token account里面的交易数量
	var tokenTxCount = 0
	tokenAccounts, err := GetSPLTokenAccounts(rpcClient, nativeAccount)
	if err != nil && err.Error() != "can not find any token account" {
		return 0, err
	}
	limit = 1000
	for _, account := range tokenAccounts {
		tokenTransactions, err := rpcClient.GetSignaturesForAddressWithOpts(context.Background(), account, &rpc.GetSignaturesForAddressOpts{
			Limit:      &limit,
			Commitment: rpc.CommitmentFinalized,
		})
		if err != nil {
			return 0, err
		}
		tokenTxCount += len(tokenTransactions)
	}
	transactionCount := (len(nativeTransactions)) + tokenTxCount
	return transactionCount, nil
}

func GetTransactionByTxId(ctx context.Context, rpcClient *rpc.Client, txId solana.Signature) (*solana.Transaction, error) {
	var maxSupportVersion uint64 = 0
	out, err := rpcClient.GetTransaction(
		ctx,
		txId,
		&rpc.GetTransactionOpts{
			MaxSupportedTransactionVersion: &maxSupportVersion,
			Commitment:                     rpc.CommitmentFinalized,
			Encoding:                       solana.EncodingBase64,
		},
	)
	if err != nil {
		return nil, err
	}
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func GetTransactionResultByTxId(ctx context.Context, rpcClient *rpc.Client, txId solana.Signature) (*rpc.GetTransactionResult, error) {
	var maxSupportVersion uint64 = 0
	tr, err := rpcClient.GetTransaction(
		ctx,
		txId,
		&rpc.GetTransactionOpts{
			MaxSupportedTransactionVersion: &maxSupportVersion,
			Commitment:                     rpc.CommitmentFinalized,
			Encoding:                       solana.EncodingBase64,
		},
	)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

func GetParsedTransactionById(ctx context.Context, rpcClient *rpc.Client, txId solana.Signature) (*rpc.GetParsedTransactionResult, error) {
	out, err := rpcClient.GetParsedTransaction(ctx, txId, nil)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func DecodeSPLTransferBySig(rpcClient *rpc.Client, txSig solana.Signature) (*token.TransferChecked, error) {
	out, err := rpcClient.GetTransaction(
		context.TODO(),
		txSig,
		&rpc.GetTransactionOpts{
			Encoding:   solana.EncodingBase64,
			Commitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		return nil, err
	}
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(out.Transaction.GetBinary()))
	if err != nil {
		return nil, err
	}
	return DecodeSPLTokenTransfer(tx)
}

func DecodeSPLTokenTransfer(tx *solana.Transaction) (*token.TransferChecked, error) {
	// 依次解析命令，解析出的TransferChecked指令为止
	for _, inst := range tx.Message.Instructions {
		accounts, err := inst.ResolveInstructionAccounts(&tx.Message)
		if err != nil {
			continue
		}
		decodedInst, err := token.DecodeInstruction(accounts, inst.Data)
		if err != nil {
			continue
		}
		if transferInfo, ok := decodedInst.Impl.(*token.TransferChecked); ok {
			fmt.Printf("amount %d\n", *transferInfo.Amount)
			fmt.Printf("decimals %d\n", *transferInfo.Decimals)
			return transferInfo, nil
		}
	}
	return nil, errors.New("can not find any transfer checked instruction in transaction")
}

// TransferSPLTokenWithPriorityFee 发送SPLToken
func TransferSPLTokenWithPriorityFee(rpcClient *rpc.Client,
	tokenMintAccount solana.PublicKey,
	fromPrivateKey solana.PrivateKey,
	toNativeAccount solana.PublicKey,
	amount uint64,
	decimals uint8,
	memo string,
	priorityFee entity.PriorityFee,
	computeUnitDetail entity.ComputeUnitDetail) (*solana.Signature, error) {

	ctx := context.Background()

	fromNativeAccount := fromPrivateKey.PublicKey()
	recent, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(fromNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, err
	}

	toTokenAccountExist := true
	toTokenAccount, _, err := solana.FindAssociatedTokenAddress(toNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, err
	}
	if _, err := GetNativeAccountByTokenAccount(rpcClient, toTokenAccount); err != nil {
		toTokenAccountExist = false
	}

	computePrice := priorityFee.PerComputeUnit.High
	if computePrice > priorityFee.PerComputeUnit.Medium*2 {
		computePrice = priorityFee.PerComputeUnit.Medium * 2
	}
	computeUnitLimit := uint32(0)
	var instructions []solana.Instruction
	if !toTokenAccountExist {
		associated := associatedtokenaccount.NewCreateInstruction(fromNativeAccount, toNativeAccount, tokenMintAccount).Build()
		instructions = append(instructions, associated)
		computeUnitLimit += uint32(computeUnitDetail.AssociatedAccount)
	}

	transferInst := token.NewTransferCheckedInstructionBuilder().
		SetAmount(amount).
		SetSourceAccount(fromTokenAccount).
		SetDestinationAccount(toTokenAccount).
		SetDecimals(decimals).
		SetMintAccount(tokenMintAccount).
		SetOwnerAccount(fromNativeAccount).
		Build()

	instructions = append(instructions, transferInst)
	computeUnitLimit += uint32(computeUnitDetail.TransferChecked)
	if memo != "" {
		programID := solana.MemoProgramID
		accounts := []*solana.AccountMeta{solana.Meta(fromNativeAccount).SIGNER().WRITE()}
		data := []byte(memo)
		memoInst := solana.NewInstruction(programID, accounts, data)
		instructions = append(instructions, memoInst)
		computeUnitLimit += uint32(computeUnitDetail.Memo)
	}
	computeUnitLimit = uint32(float64(computeUnitLimit) * 1.2)
	computeBudgetLimitInst := computebudget.NewSetComputeUnitLimitInstructionBuilder().SetUnits(computeUnitLimit).Build()
	computeBudgetPriceInst := computebudget.NewSetComputeUnitPriceInstructionBuilder().SetMicroLamports(computePrice).Build()
	instructions = append([]solana.Instruction{computeBudgetLimitInst}, instructions...)
	instructions = append([]solana.Instruction{computeBudgetPriceInst}, instructions...)

	tx, err := solana.NewTransaction(instructions, recent.Value.Blockhash, solana.TransactionPayer(fromNativeAccount))
	if err != nil {
		return nil, err
	}

	if _, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativeAccount.Equals(key) {
				return &fromPrivateKey
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	// 模拟发送交易
	simulateResult, err := rpcClient.SimulateTransactionWithOpts(ctx, tx, &rpc.SimulateTransactionOpts{
		ReplaceRecentBlockhash: true,
	})
	if err != nil {
		return nil, fmt.Errorf("simulate transaction error %v", err)
	}
	if simulateResult.Value.Err != nil {
		return nil, fmt.Errorf("simulate transaction failed %v", simulateResult.Value.Err)
	}

	txId, err := rpcClient.SendTransaction(ctx, tx)
	return &txId, err
}

// TransferSPLToken 发送SPLToken
func TransferSPLToken(rpcClient *rpc.Client,
	tokenMintAccount solana.PublicKey,
	fromPrivateKey solana.PrivateKey,
	toNativeAccount solana.PublicKey,
	amount uint64,
	decimals uint8,
	memo string) (*solana.Signature, error) {

	ctx := context.Background()

	fromNativeAccount := fromPrivateKey.PublicKey()
	recent, err := rpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(fromNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, err
	}

	toTokenAccountExist := true
	toTokenAccount, _, err := solana.FindAssociatedTokenAddress(toNativeAccount, tokenMintAccount)
	if err != nil {
		return nil, err
	}
	if _, err := GetNativeAccountByTokenAccount(rpcClient, toTokenAccount); err != nil {
		toTokenAccountExist = false
	}

	var instructions []solana.Instruction
	if !toTokenAccountExist {
		associated := associatedtokenaccount.NewCreateInstruction(fromNativeAccount, toNativeAccount, tokenMintAccount).Build()
		instructions = append(instructions, associated)
	}

	transferInst := token.NewTransferCheckedInstructionBuilder().
		SetAmount(amount).
		SetSourceAccount(fromTokenAccount).
		SetDestinationAccount(toTokenAccount).
		SetDecimals(decimals).
		SetMintAccount(tokenMintAccount).
		SetOwnerAccount(fromNativeAccount).
		Build()

	instructions = append(instructions, transferInst)
	if memo != "" {
		programID := solana.MemoProgramID
		accounts := []*solana.AccountMeta{solana.Meta(fromNativeAccount).SIGNER().WRITE()}
		data := []byte(memo)
		memoInst := solana.NewInstruction(programID, accounts, data)
		instructions = append(instructions, memoInst)
	}

	tx, err := solana.NewTransaction(instructions, recent.Value.Blockhash, solana.TransactionPayer(fromNativeAccount))
	if err != nil {
		return nil, err
	}

	if _, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativeAccount.Equals(key) {
				return &fromPrivateKey
			}
			return nil
		},
	); err != nil {
		return nil, err
	}

	// 模拟发送交易
	simulateResult, err := rpcClient.SimulateTransactionWithOpts(ctx, tx, &rpc.SimulateTransactionOpts{
		ReplaceRecentBlockhash: true,
	})
	if err != nil {
		return nil, fmt.Errorf("simulate transaction error %v", err)
	}
	if simulateResult.Value.Err != nil {
		return nil, fmt.Errorf("simulate transaction failed %v", simulateResult.Value.Err)
	}

	txId, err := rpcClient.SendTransaction(ctx, tx)
	return &txId, err
}

type QnEstimatePriorityFeesParam struct {
	LastNBlocks int    `json:"last_n_blocks"`
	Account     string `json:"account"`
}

type QnRequestBody struct {
	JSONRPC string                      `json:"jsonrpc"`
	ID      int                         `json:"id"`
	Method  string                      `json:"method"`
	Params  QnEstimatePriorityFeesParam `json:"params"`
}

// ResponseBody 用于存储响应体
type QnEstimatePriorityFee struct {
	JSONRPC string `json:"jsonrpc"`
	Result  struct {
		Context struct {
			Slot int `json:"slot"`
		} `json:"context"`
		PerComputeUnit struct {
			Extreme     int `json:"extreme"`
			High        int `json:"high"`
			Low         int `json:"low"`
			Medium      int `json:"medium"`
			Percentiles struct {
				P50  int `json:"50"`
				P55  int `json:"55"`
				P60  int `json:"60"`
				P65  int `json:"65"`
				P70  int `json:"70"`
				P75  int `json:"75"`
				P80  int `json:"80"`
				P85  int `json:"85"`
				P90  int `json:"90"`
				P95  int `json:"95"`
				P100 int `json:"100"`
			}
		} `json:"per_compute_unit"`

		PerTransaction struct {
			Extreme     int `json:"extreme"`
			High        int `json:"high"`
			Low         int `json:"low"`
			Medium      int `json:"medium"`
			Percentiles struct {
				P50  int `json:"50"`
				P55  int `json:"55"`
				P60  int `json:"60"`
				P65  int `json:"65"`
				P70  int `json:"70"`
				P75  int `json:"75"`
				P80  int `json:"80"`
				P85  int `json:"85"`
				P90  int `json:"90"`
				P95  int `json:"95"`
				P100 int `json:"100"`
			}
		} `json:"per_transaction"`
	} `json:"result"`
	ID int `json:"id"`
}

func GetQnEstimatePriorityFees(url string, lastBlocks int, programId solana.PublicKey) (*QnEstimatePriorityFee, error) {
	params := QnEstimatePriorityFeesParam{
		LastNBlocks: lastBlocks,
		Account:     programId.String(),
	}
	requestBody := QnRequestBody{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "qn_estimatePriorityFees",
		Params:  params,
	}
	// 将请求体序列化为 JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %v", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	// 设置请求头
	req.Header.Set("x-qn-api-version", "1")
	req.Header.Set("Content-Type", "application/json")

	// 发送 HTTP 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// 解析响应为 JSON
	var estimatePriorityFee QnEstimatePriorityFee
	if err := json.Unmarshal(body, &estimatePriorityFee); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}
	return &estimatePriorityFee, nil
}

func TransferSPLTokenWithFeePayer(rpcClient *rpc.Client, wssClient *ws.Client, tokenMintAccount solana.PublicKey,
	fromPrivKey solana.PrivateKey, feePayerPrivKey solana.PrivateKey, fromTokenPubKey solana.PublicKey, toTokenPubKey solana.PublicKey, amount uint64, decimals uint8) (*solana.Signature, error) {

	ctx := context.Background()

	// TODO: 从数据库当中读取
	//srcPrivKey := solana.MustPrivateKeyFromBase58("3coeLXLqdWi9kwSGDrNZZWf5bv6mBW2Zmd8gBSYJpqiGjy73cpmFYHM33Nn5Sxr2i2jJcE4xRXqDuDPTRtn3iHsz")
	fromNativePubKey := fromPrivKey.PublicKey()
	feePayerNativePubKey := feePayerPrivKey.PublicKey()

	recent, err := rpcClient.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	transferInst := token.NewTransferCheckedInstructionBuilder().
		SetAmount(amount).
		SetDecimals(decimals).
		SetSourceAccount(fromTokenPubKey).
		SetMintAccount(tokenMintAccount).
		SetDestinationAccount(toTokenPubKey).
		SetOwnerAccount(fromNativePubKey).
		Build()

	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			transferInst,
		},
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayerNativePubKey),
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativePubKey.Equals(key) {
				return &fromPrivKey
			}
			if feePayerNativePubKey.Equals(key) {
				return &feePayerPrivKey
			}
			return nil
		},
	)

	txSig, err := confirm.SendAndConfirmTransaction(ctx, rpcClient, wssClient, tx)
	return &txSig, err
}

// MultiTransferSPLToken 发送SPLToken
func MultiTransferSPLToken(rpcClient *rpc.Client, wssClient *ws.Client, tokenMintAccount solana.PublicKey,
	fromPrivKey solana.PrivateKey, fromTokenPubKey solana.PublicKey, toTokenPubKey []solana.PublicKey, amount uint64, decimals uint8) (*solana.Signature, error) {

	ctx := context.Background()

	// TODO: 从数据库当中读取
	fromNativePubKey := fromPrivKey.PublicKey()

	recent, err := rpcClient.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	var instList []solana.Instruction
	for _, to := range toTokenPubKey {
		transferInst := token.NewTransferCheckedInstructionBuilder().
			SetAmount(amount).
			SetDecimals(decimals).
			SetSourceAccount(fromTokenPubKey).
			SetMintAccount(tokenMintAccount).
			SetDestinationAccount(to).
			SetOwnerAccount(fromNativePubKey).
			Build()
		instList = append(instList, transferInst)
	}

	tx, err := solana.NewTransaction(
		instList,
		recent.Value.Blockhash,
		solana.TransactionPayer(fromNativePubKey),
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativePubKey.Equals(key) {
				return &fromPrivKey
			}
			return nil
		},
	)

	txSig, err := confirm.SendAndConfirmTransaction(ctx, rpcClient, wssClient, tx)
	return &txSig, err
}

// MultiTransferSPLTokenWithFeePayer 发送SPLToken
func MultiTransferSPLTokenWithFeePayer(rpcClient *rpc.Client, wssClient *ws.Client, tokenMintAccount solana.PublicKey,
	fromPrivKey solana.PrivateKey, feePayerPrivKey solana.PrivateKey, fromTokenPubKey solana.PublicKey,
	toTokenPubKey []solana.PublicKey, amount uint64, decimals uint8) (*solana.Signature, error) {

	ctx := context.Background()

	// TODO: 从数据库当中读取
	fromNativePubKey := fromPrivKey.PublicKey()
	feePayerNativePubKey := feePayerPrivKey.PublicKey()

	recent, err := rpcClient.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	var instList []solana.Instruction
	for _, to := range toTokenPubKey {
		transferInst := token.NewTransferCheckedInstructionBuilder().
			SetAmount(amount).
			SetDecimals(decimals).
			SetSourceAccount(fromTokenPubKey).
			SetMintAccount(tokenMintAccount).
			SetDestinationAccount(to).
			SetOwnerAccount(fromNativePubKey).
			Build()
		instList = append(instList, transferInst)
	}

	tx, err := solana.NewTransaction(
		instList,
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayerNativePubKey),
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativePubKey.Equals(key) {
				return &fromPrivKey
			}
			if feePayerNativePubKey.Equals(key) {
				return &feePayerPrivKey
			}
			return nil
		},
	)

	txSig, err := confirm.SendAndConfirmTransaction(ctx, rpcClient, wssClient, tx)
	return &txSig, err
}

// TODO: 手续费优化
func MultiTransferSPLRewardToken(ctx context.Context,
	rpcClient *rpc.Client,
	fromPrivKey solana.PrivateKey,
	toTokenAccounts []solana.PublicKey,
	tokenMintAccounts []solana.PublicKey,
	rewardAmounts []uint64,
	decimals []uint8) (*solana.Signature, uint64, error) {

	fromNativePubKey := fromPrivKey.PublicKey()

	recent, err := rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, 0, err
	}

	var instList []solana.Instruction
	computeBudgetPriceInst := computebudget.NewSetComputeUnitPriceInstructionBuilder().
		SetMicroLamports(1000).
		Build()
	computeBudgetLimitInst := computebudget.NewSetComputeUnitLimitInstructionBuilder().
		SetUnits(600000).
		Build()
	instList = append(instList, computeBudgetPriceInst)
	instList = append(instList, computeBudgetLimitInst)

	for index, to := range toTokenAccounts {
		fromTokenAccount, _, _ := solana.FindAssociatedTokenAddress(fromNativePubKey, tokenMintAccounts[index])
		transferInst := token.NewTransferCheckedInstructionBuilder().
			SetAmount(rewardAmounts[index]).
			SetDecimals(decimals[index]).
			SetSourceAccount(fromTokenAccount).
			SetMintAccount(tokenMintAccounts[index]).
			SetDestinationAccount(to).
			SetOwnerAccount(fromNativePubKey).
			Build()
		instList = append(instList, transferInst)
	}

	tx, err := solana.NewTransaction(
		instList,
		recent.Value.Blockhash,
		solana.TransactionPayer(fromNativePubKey),
	)
	if err != nil {
		return nil, 0, err
	}
	messageBin, err := tx.Message.MarshalBinary()
	if err != nil {
		return nil, 0, err
	}
	// 评估手续费
	base64EncodeMsg := base64.StdEncoding.EncodeToString(messageBin)
	fee, err := rpcClient.GetFeeForMessage(ctx, base64EncodeMsg, rpc.CommitmentConfirmed)
	if err != nil {
		return nil, 0, err
	}
	if fee == nil || fee.Value == nil {
		return nil, 0, err
	}
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if fromNativePubKey.Equals(key) {
				return &fromPrivKey
			}
			return nil
		},
	)
	txId, err := rpcClient.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
		SkipPreflight: true,
	})
	return &txId, *fee.Value, err
}

func CalcGasFee(sigNum, lamPerSig, computeUnitPrice, computeUnitLimit uint64, computeUnitSet bool) uint64 {
	// 评估手续费的时候，如果没有设置computeUnit，则按照 200_000算
	var unitLimit uint64 = 0
	if computeUnitSet {
		unitLimit = computeUnitLimit
	} else {
		unitLimit = DefaultComputeUnitLimit
	}
	sigFee := sigNum * lamPerSig
	computeUnitsDFee := (float64(computeUnitPrice) * 1000 / float64(solana.LAMPORTS_PER_SOL)) * float64(unitLimit)
	// 使用 math.Ceil 对总费用进行向上取整
	return uint64(math.Ceil(float64(sigFee) + computeUnitsDFee))
}

// GetInstructionDiscriminator 计算指令的 discriminator
func GetInstructionDiscriminator(instructionName string) []byte {
	hash := sha256.Sum256([]byte("global:" + instructionName))
	return hash[:8]
}

// FindPDA 查找PDA账户
func FindPDA(programID solana.PublicKey, seeds []byte, additionalSeeds ...[]byte) (solana.PublicKey, uint8, error) {
	allSeeds := [][]byte{seeds}
	allSeeds = append(allSeeds, additionalSeeds...)
	return solana.FindProgramAddress(allSeeds, programID)
}

// GenerateRandomSolanaTxID 生成随机的 Solana 交易 ID（Signature）
func GenerateRandomSolanaTxID() (solana.Signature, error) {
	// Solana 交易签名是 64 字节的
	signatureBytes := make([]byte, 64)

	// 使用加密安全的随机数生成器填充字节
	_, err := rand.Read(signatureBytes)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// 创建 solana.Signature
	signature := solana.SignatureFromBytes(signatureBytes)
	return signature, nil
}

// GenerateRandomSolanaTxIDString 生成随机的 Solana 交易 ID 字符串
func GenerateRandomSolanaTxIDString() (string, error) {
	signature, err := GenerateRandomSolanaTxID()
	if err != nil {
		return "", err
	}
	return signature.String(), nil
}

func SaveSolanaPrivateKeyToJSON(base58PrivateKey, filename, filePath string) error {
	// Decode Base58 private key to byte array
	privateKeyBytes, err := base58.Decode(base58PrivateKey)
	if err != nil {
		return fmt.Errorf("invalid Base58 private key: %v", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filePath, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Build full file path
	fullPath := filepath.Join(filePath, filename)

	// Create and write to JSON file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Set up JSON encoder
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	// Encode to JSON and write to file
	if err := encoder.Encode(privateKeyBytes); err != nil {
		return fmt.Errorf("JSON encoding failed: %v", err)
	}

	fmt.Printf("Private key successfully saved to: %s\n", fullPath)
	return nil
}

// FilterAndTranslateSOLError 过滤并翻译SOL错误信息为英文提示
func FilterAndTranslateSOLError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// 1. RPC节点网络错误
	if strings.Contains(errStr, "status code: 50") {
		if strings.Contains(errStr, "status code: 502") {
			return "Network service error, please try again later"
		} else if strings.Contains(errStr, "status code: 503") {
			return "Service temporarily unavailable, please try again later"
		} else if strings.Contains(errStr, "status code: 504") {
			return "Request timeout, please try again later"
		}
		return "Network connection error, please try again later"
	}

	// 2. 超时错误
	if strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(errStr, "Client.Timeout exceeded") {
		return "Request timeout, please check your network connection"
	}

	// 3. 交易模拟失败
	if strings.Contains(errStr, "Transaction simulation failed") {
		// 余额不足
		if strings.Contains(strings.ToLower(errStr), "insufficient funds") {
			return "Insufficient account balance, please ensure you have enough SOL for transaction fees and rent"
		}

		// 区块哈希过期
		if strings.Contains(errStr, "Blockhash not found") {
			return "Transaction expired, please refresh and try again"
		}

		// 交易已处理
		if strings.Contains(errStr, "already been processed") {
			return "Transaction already processed successfully, no need to resubmit"
		}

		// 自定义程序错误 - 修改这里的提示
		if strings.Contains(errStr, "custom program error") {
			// 提取错误码
			re := regexp.MustCompile(`custom program error:\s*(0x[0-9a-fA-F]+|\d+)`)
			if match := re.FindStringSubmatch(errStr); len(match) > 1 {
				errorCode := match[1]
				// 根据常见错误码提供更详细的提示
				switch errorCode {
				case "0x1900", "6400": // 十进制6400 = 十六进制0x1900
					return "Wallet compatibility issue. Please try using a different wallet or contact our support team for assistance."
				case "0x0": // 常见错误码
					return "Wallet compatibility issue. Please try using a different wallet or contact our support team for assistance."
				default:
					return fmt.Sprintf("Wallet compatibility issue. Please try using a different wallet or contact our support team for assistance. (Error code: %s)", errorCode)
				}
			}
			return "Wallet compatibility issue. Please try using a different wallet or contact our support team for assistance."
		}

		// 其他模拟失败
		return "Transaction validation failed, please check transaction data"
	}

	// 4. RPC连接错误
	if strings.Contains(errStr, "rpc call sendTransaction") {
		if strings.Contains(errStr, "Post") && strings.Contains(errStr, "EOF") {
			return "Network connection interrupted, please try again later"
		}
		return "Network request failed, please check your connection"
	}

	// 5. 通用错误提示
	return "Failed to send transaction, please try again later"
}

// TransferCheckedParams 定义返回的TransferChecked指令参数结构体
type TransferCheckedParams struct {
	SourceAccount string // 源账户
	DestAccount   string // 目标账户
	MintAccount   string // Mint账户
	OwnerAccount  string // Owner账户
	Amount        uint64 // 转账金额
	Decimals      uint8  // Token小数位数
}

// ParseMarketBuyTx 优化版核心函数
// 参数：
//   - rpcClient: RPC客户端
//   - txSig: 交易ID
//
// 说明:
//   - 该函数仅限主网使用，专门用来解析主网购买token的交易
//   - 修改为通用版本，不再硬编码特定市场账户
//
// 返回：解析后的Solana交易结构体 / 错误信息
func ParseMarketBuyTx(rpcClient *rpc.Client, txSig solana.Signature) (*entity.DecodedSolanaTransaction, error) {
	sourceAccount := solana.MPK("GjkvqFpZ5gqbzYEUAGsn5ozmFgM52JJDgso426DiLXbQ")
	tokenMint := solana.MPK("ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH")

	// 1. 获取交易数据（使用外部传入的RPC客户端）
	txResp, err := getTransaction(rpcClient, txSig)
	if err != nil {
		return nil, fmt.Errorf("获取交易失败: %w", err)
	}

	// 2. 解析交易二进制数据
	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txResp.Transaction.GetBinary()))
	if err != nil {
		return nil, fmt.Errorf("解析交易失败: %w", err)
	}

	// 3. 预解析目标账户公钥
	// 构建完整账户列表（包含查找表账户）
	fullAccountList := buildFullAccountList(tx, txResp)

	// 4. 直接解析Inner Instructions，查找符合条件的TransferChecked指令
	// 不再检查是否包含特定市场账户，因为不同DEX使用不同的账户结构
	transferCheckedInst, err := parseInnerInstructionsForTransferChecked(txResp, fullAccountList, sourceAccount, tokenMint)
	if err != nil {
		return nil, fmt.Errorf("解析Inner Instructions失败: %w", err)
	}
	if transferCheckedInst != nil {
		// 创建并返回DecodedSolanaTransaction
		decodedTx := &entity.DecodedSolanaTransaction{
			TxID:                        txSig,
			FromNativeAccount:           tx.Message.AccountKeys[0], // 第一个账户是手续费支付者，与DecodeSolanaTransaction保持一致
			FromTokenAccount:            solana.MustPublicKeyFromBase58(transferCheckedInst.SourceAccount),
			FeePayer:                    tx.Message.AccountKeys[0], // 第一个账户通常是手续费支付者
			RefBlockHash:                tx.Message.RecentBlockhash,
			Accounts:                    fullAccountList,
			Signatures:                  tx.Signatures,
			TransferInstructions:        []entity.DecodedSolTransferInst{},
			TransferCheckedInstructions: []entity.DecodedSolTransferCheckedInst{},
			ComputeUnitPrice:            0,
			ComputeUnitLimit:            0,
			EstimateFee:                 0,
		}

		// 添加TransferChecked指令
		transferChecked := entity.DecodedSolTransferCheckedInst{
			FromTokenAccount:   solana.MustPublicKeyFromBase58(transferCheckedInst.SourceAccount),
			FromNativeAccount:  solana.MustPublicKeyFromBase58(transferCheckedInst.OwnerAccount), // token account的owner
			ToTokenAccount:     solana.MustPublicKeyFromBase58(transferCheckedInst.DestAccount),
			ToNativeAccount:    solana.PublicKey{},                                               // 需要从token account获取owner
			OwnerNativeAccount: solana.MustPublicKeyFromBase58(transferCheckedInst.OwnerAccount), // 指令中的owner账户
			TokenMintAccount:   solana.MustPublicKeyFromBase58(transferCheckedInst.MintAccount),
			Amount:             transferCheckedInst.Amount,
			Decimals:           transferCheckedInst.Decimals,
		}
		decodedTx.TransferCheckedInstructions = append(decodedTx.TransferCheckedInstructions, transferChecked)

		// 尝试获取ToNativeAccount（目标token account的owner）
		// 首先检查DestAccount是否是原生账户（对于Jupiter/TransitSwap交易，DestAccount可能是用户原生账户）
		destAccount := solana.MustPublicKeyFromBase58(transferCheckedInst.DestAccount)

		// 检查是否是系统账户或已知的程序ID
		isSystemAccount := destAccount.Equals(solana.SystemProgramID) ||
			destAccount.Equals(solana.TokenProgramID) ||
			destAccount.Equals(solana.MPK("GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL")) ||
			destAccount.Equals(solana.MPK("8AgyxWiUW4Wczmqsh89vz6cjkmkPvJbEeU7WS51jkGWv"))

		if !isSystemAccount {
			// 尝试获取token账户的owner
			if owner, err := GetNativeAccountByTokenAccount(rpcClient, destAccount); err == nil && owner != nil {
				decodedTx.TransferCheckedInstructions[0].ToNativeAccount = *owner
			} else {
				// 如果获取失败，假设DestAccount已经是原生账户
				decodedTx.TransferCheckedInstructions[0].ToNativeAccount = destAccount
			}
		}

		return decodedTx, nil // 找到符合条件的指令，直接返回
	}

	return nil, fmt.Errorf("未找到符合条件的TransferChecked指令")
}

// 构建完整账户列表（包含查找表账户）
func buildFullAccountList(tx *solana.Transaction, txResp *rpc.GetTransactionResult) []solana.PublicKey {
	fullList := make([]solana.PublicKey, len(tx.Message.AccountKeys))
	copy(fullList, tx.Message.AccountKeys)

	if txResp.Meta != nil {
		// 追加查找表的ReadOnly/Writable账户
		for _, addr := range txResp.Meta.LoadedAddresses.ReadOnly {
			fullList = append(fullList, addr)
		}
		for _, addr := range txResp.Meta.LoadedAddresses.Writable {
			fullList = append(fullList, addr)
		}
	}
	return fullList
}

// 检查指令是否包含目标Input Account（核心筛选条件）
func checkInstructionHasTargetAccount(tx *solana.Transaction, inst *solana.CompiledInstruction, targetPubKey solana.PublicKey) bool {
	for _, accIndex := range inst.Accounts {
		if int(accIndex) >= len(tx.Message.AccountKeys) {
			continue
		}
		if tx.Message.AccountKeys[accIndex].Equals(targetPubKey) {
			return true
		}
	}
	return false
}

// 获取交易详情（复用外部RPC客户端）
func getTransaction(client *rpc.Client, sig solana.Signature) (*rpc.GetTransactionResult, error) {
	version := uint64(0)
	return client.GetTransaction(
		context.TODO(),
		sig,
		&rpc.GetTransactionOpts{
			MaxSupportedTransactionVersion: &version,
			Encoding:                       solana.EncodingBase64,
			Commitment:                     rpc.CommitmentFinalized,
		},
	)
}

// 解析Inner Instructions，筛选source account匹配的TransferChecked指令
func parseInnerInstructionsForTransferChecked(txResp *rpc.GetTransactionResult, fullAccountList []solana.PublicKey, source, mint solana.PublicKey) (*TransferCheckedParams, error) {
	if txResp.Meta == nil {
		return nil, nil // 无Meta数据，非错误，返回nil
	}

	// 辅助函数：从token balances中提取金额
	extractAmountFromBalances := func() (uint64, uint8, solana.PublicKey, bool) {
		if len(txResp.Meta.PreTokenBalances) == 0 || len(txResp.Meta.PostTokenBalances) == 0 {
			fmt.Printf("调试: 无token balances数据\n")
			return 0, 0, solana.PublicKey{}, false
		}

		// 查找市场账户的余额变化（token账户的owner）
		// source是token账户，但我们需要找它的owner（市场账户）的余额
		var marketPreBalance, marketPostBalance uint64
		var decimals uint8
		var userAccount solana.PublicKey

		mintStr := mint.String()
		// 市场账户是token账户的owner，对于我们的case是GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL
		marketAccount := solana.MPK("GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL")
		marketStr := marketAccount.String()

		fmt.Printf("调试: 查找市场账户=%s, mint=%s的余额变化\n", marketStr, mintStr)
		fmt.Printf("调试: PreTokenBalances数量=%d, PostTokenBalances数量=%d\n",
			len(txResp.Meta.PreTokenBalances), len(txResp.Meta.PostTokenBalances))

		for _, pre := range txResp.Meta.PreTokenBalances {
			fmt.Printf("调试: Pre - Owner:%s, Mint:%s, Amount:%s, Decimals:%d\n",
				pre.Owner, pre.Mint, pre.UiTokenAmount.Amount, pre.UiTokenAmount.Decimals)
			if pre.Mint.String() == mintStr && pre.Owner.String() == marketStr {
				// 解析余额字符串
				var amount uint64
				fmt.Sscanf(pre.UiTokenAmount.Amount, "%d", &amount)
				marketPreBalance = amount
				decimals = uint8(pre.UiTokenAmount.Decimals)
				fmt.Printf("调试: 找到市场 Pre余额: %d, decimals: %d\n", amount, decimals)
			}
		}

		for _, post := range txResp.Meta.PostTokenBalances {
			fmt.Printf("调试: Post - Owner:%s, Mint:%s, Amount:%s, Decimals:%d\n",
				post.Owner, post.Mint, post.UiTokenAmount.Amount, post.UiTokenAmount.Decimals)
			if post.Mint.String() == mintStr {
				if post.Owner.String() == marketStr {
					// 解析余额字符串
					var amount uint64
					fmt.Sscanf(post.UiTokenAmount.Amount, "%d", &amount)
					marketPostBalance = amount
					fmt.Printf("调试: 找到市场 Post余额: %d\n", amount)
				} else if post.Owner.String() != marketStr {
					// 这可能是用户账户（接收方）
					userAccount = *post.Owner
					fmt.Printf("调试: 找到用户账户: %s, 余额: %s\n", userAccount, post.UiTokenAmount.Amount)
				}
			}
		}

		fmt.Printf("调试: marketPreBalance=%d, marketPostBalance=%d, userAccount=%s\n",
			marketPreBalance, marketPostBalance, userAccount)

		if marketPreBalance > marketPostBalance && !userAccount.IsZero() {
			// 市场账户余额减少，表示转出给用户
			amount := marketPreBalance - marketPostBalance
			fmt.Printf("调试: 计算金额: %d - %d = %d\n", marketPreBalance, marketPostBalance, amount)
			return amount, decimals, userAccount, true
		}

		fmt.Printf("调试: 无法提取金额\n")
		return 0, 0, solana.PublicKey{}, false
	}

	if len(txResp.Meta.InnerInstructions) == 0 {
		// 即使没有inner instructions，也可以尝试从balances中提取信息
		if amount, decimals, destAccount, ok := extractAmountFromBalances(); ok {
			return &TransferCheckedParams{
				SourceAccount: source.String(),
				DestAccount:   destAccount.String(),
				MintAccount:   mint.String(),
				OwnerAccount:  "GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL", // 默认
				Amount:        amount,
				Decimals:      decimals,
			}, nil
		}
		return nil, nil
	}

	tokenProgramID := solana.TokenProgramID
	jupiterWrapperProgramID := solana.MPK("GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL")
	transitSwapProgramID := solana.MPK("8AgyxWiUW4Wczmqsh89vz6cjkmkPvJbEeU7WS51jkGWv")

	for _, innerInstGroup := range txResp.Meta.InnerInstructions {
		for _, innerInst := range innerInstGroup.Instructions {
			// 解析Inner Instruction的Program ID
			var progID solana.PublicKey
			if int(innerInst.ProgramIDIndex) >= len(fullAccountList) {
				continue // 无法解析程序ID，跳过
			} else {
				progID = fullAccountList[innerInst.ProgramIDIndex]
			}

			// 检查是否是Token程序、Jupiter包装程序或TransitSwap程序
			isTokenProgram := progID.Equals(tokenProgramID)
			isJupiterWrapper := progID.Equals(jupiterWrapperProgramID)
			isTransitSwap := progID.Equals(transitSwapProgramID)

			if !isTokenProgram && !isJupiterWrapper && !isTransitSwap {
				continue // 不是目标程序，跳过
			}

			// 检查指令是否包含目标source账户
			containsSource := false
			for _, accIndex := range innerInst.Accounts {
				if int(accIndex) < len(fullAccountList) {
					account := fullAccountList[accIndex]
					if account.Equals(source) {
						containsSource = true
						break
					}
				}
			}

			if !containsSource {
				continue // 不包含目标source账户，跳过
			}

			// 对于Token程序指令，尝试解码
			if isTokenProgram {
				params, err := decodeTokenInstructionForTransferChecked(innerInst, fullAccountList, source, mint)
				if err != nil {
					fmt.Printf("解码Token指令失败: %v\n", err)
					continue
				}
				if params != nil {
					return params, nil // 找到符合条件的Token指令
				}
			}

			// 对于Jupiter包装程序或TransitSwap程序，尝试提取TransferChecked信息
			// 首先尝试直接解码指令数据（可能包含TransferChecked格式的数据）
			if isJupiterWrapper || isTransitSwap {
				// 尝试直接解码指令数据
				params, err := decodeTokenInstructionForTransferChecked(innerInst, fullAccountList, source, mint)
				if err == nil && params != nil {
					return params, nil // 成功解码
				}

				// 如果直接解码失败，尝试从账户列表中提取信息
				// 检查指令是否包含Token程序账户
				// 对于Jupiter: Token程序通常是账户[0]
				// 对于TransitSwap: Token程序是账户[6] (从测试输出看)
				var tokenProgramIndex uint16
				if isJupiterWrapper && len(innerInst.Accounts) >= 1 {
					tokenProgramIndex = innerInst.Accounts[0]
				} else if isTransitSwap && len(innerInst.Accounts) >= 7 {
					tokenProgramIndex = innerInst.Accounts[6] // TransitSwap的Token程序在索引6
				} else {
					continue
				}

				if int(tokenProgramIndex) < len(fullAccountList) {
					tokenProgram := fullAccountList[tokenProgramIndex]
					if tokenProgram.Equals(tokenProgramID) {
						// 这可能是包装的Token指令
						// 尝试提取destination账户
						var destAccount solana.PublicKey
						var sourceAccount solana.PublicKey

						if isJupiterWrapper && len(innerInst.Accounts) >= 4 {
							// Jupiter结构: [TokenProgram, MarketAccount, DestAccount, SourceAccount]
							destIndex := innerInst.Accounts[2]
							sourceIndex := innerInst.Accounts[3]
							if int(destIndex) < len(fullAccountList) && int(sourceIndex) < len(fullAccountList) {
								destAccount = fullAccountList[destIndex]
								sourceAccount = fullAccountList[sourceIndex]
							}
						} else if isTransitSwap && len(innerInst.Accounts) >= 13 {
							// TransitSwap结构更复杂
							// 从测试看: 账户[5]是dest, 账户[10]是source
							destIndex := innerInst.Accounts[5]
							sourceIndex := innerInst.Accounts[10]
							if int(destIndex) < len(fullAccountList) && int(sourceIndex) < len(fullAccountList) {
								destAccount = fullAccountList[destIndex]
								sourceAccount = fullAccountList[sourceIndex]
							}
						}

						if !destAccount.IsZero() && !sourceAccount.IsZero() {
							// 验证source账户是否匹配
							if !sourceAccount.Equals(source) {
								continue // source不匹配
							}

							// 尝试从token balances中提取金额
							amount, decimals, balanceDestAccount, ok := extractAmountFromBalances()
							if !ok {
								// 如果无法从balances提取，使用默认值
								amount = 0
								decimals = 0
							} else if !balanceDestAccount.IsZero() && !balanceDestAccount.Equals(destAccount) {
								// 如果从balances找到的dest账户不匹配，使用指令中的dest账户
								destAccount = balanceDestAccount
							}

							ownerAccount := "GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL" // 默认使用Jupiter包装程序作为owner
							if isTransitSwap {
								ownerAccount = "8AgyxWiUW4Wczmqsh89vz6cjkmkPvJbEeU7WS51jkGWv" // TransitSwap程序作为owner
							}

							return &TransferCheckedParams{
								SourceAccount: source.String(),
								DestAccount:   destAccount.String(),
								MintAccount:   mint.String(),
								OwnerAccount:  ownerAccount,
								Amount:        amount,
								Decimals:      decimals,
							}, nil
						}
					}
				}
			}
		}
	}

	return nil, nil // 未找到符合条件的指令（非错误）
}

// 清理JSON字符串的引号
func cleanJSONQuotes(data []byte) []byte {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		return data[1 : len(data)-1]
	}
	return data
}

// 判断是否为Token指令（仅关注TransferChecked）
func isTokenInstruction(opcode byte) bool {
	return opcode == 13 // 仅筛选TransferChecked指令（opcode=13）
}

// 解码Token指令，筛选source account匹配的TransferChecked
func decodeTokenInstructionForTransferChecked(inst rpc.CompiledInstruction, fullAccountList []solana.PublicKey, source, mint solana.PublicKey) (*TransferCheckedParams, error) {
	// 1. 构建AccountMeta列表
	var accounts []*solana.AccountMeta
	for _, accIndex := range inst.Accounts {
		if int(accIndex) >= len(fullAccountList) {
			continue
		}
		accounts = append(accounts, &solana.AccountMeta{
			PublicKey:  fullAccountList[accIndex],
			IsSigner:   false,
			IsWritable: false,
		})
	}

	// 2. 解析并解码指令Data
	dataJSON, err := inst.Data.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("解析Data失败: %w", err)
	}
	cleanData := cleanJSONQuotes(dataJSON)
	rawData, err := base58.Decode(string(cleanData))
	if err != nil {
		return nil, fmt.Errorf("base58解码失败: %v", err)
	}

	// 3. 首先尝试手动解析TransferChecked指令
	// TransferChecked指令格式: [opcode(13)] + [amount(8 bytes)] + [decimals(1 byte)]
	if len(rawData) >= 1 {
		opcode := rawData[0]
		fmt.Printf("调试: 指令opcode=%d (0x%x), 数据长度=%d\n", opcode, opcode, len(rawData))

		if opcode == 13 && len(rawData) >= 10 { // TransferChecked opcode
			// 解析amount (little-endian uint64)
			var amount uint64
			for i := 0; i < 8; i++ {
				amount |= uint64(rawData[1+i]) << (8 * i)
			}
			// 解析decimals
			decimals := rawData[9]

			fmt.Printf("调试: 找到TransferChecked指令, amount=%d, decimals=%d\n", amount, decimals)

			// 从账户列表中提取信息
			// TransferChecked指令的账户顺序: [source, mint, destination, owner]
			if len(accounts) >= 4 {
				sourceAccount := accounts[0].PublicKey
				mintAccount := accounts[1].PublicKey
				destAccount := accounts[2].PublicKey
				ownerAccount := accounts[3].PublicKey

				fmt.Printf("调试: 账户: source=%s, mint=%s, dest=%s, owner=%s\n",
					sourceAccount, mintAccount, destAccount, ownerAccount)

				// 检查是否匹配
				if sourceAccount.Equals(source) && mintAccount.Equals(mint) {
					return &TransferCheckedParams{
						SourceAccount: sourceAccount.String(),
						DestAccount:   destAccount.String(),
						MintAccount:   mintAccount.String(),
						OwnerAccount:  ownerAccount.String(),
						Amount:        amount,
						Decimals:      decimals,
					}, nil
				}
			}
		}
	}

	// 4. 如果手动解析失败，尝试使用库函数解码Token指令
	decodedInst, err := token.DecodeInstruction(accounts, rawData)
	if err != nil {
		return nil, fmt.Errorf("解码Token指令失败: %w", err)
	}

	// 5. 筛选TransferChecked且source account匹配的指令
	switch decoded := decodedInst.Impl.(type) {
	case *token.TransferChecked:
		// 获取Source Account并对比
		sourceAccount := decoded.GetSourceAccount().PublicKey
		if !sourceAccount.Equals(source) {
			return nil, nil // source不匹配，跳过
		}
		mintAccount := decoded.GetMintAccount().PublicKey
		if !mintAccount.Equals(mint) {
			return nil, nil // mint不匹配，跳过
		}

		// 封装返回参数
		return &TransferCheckedParams{
			SourceAccount: sourceAccount.String(),
			DestAccount:   decoded.GetDestinationAccount().PublicKey.String(),
			MintAccount:   decoded.GetMintAccount().PublicKey.String(),
			OwnerAccount:  decoded.GetOwnerAccount().PublicKey.String(),
			Amount:        *decoded.Amount,
			Decimals:      *decoded.Decimals,
		}, nil
	default:
		return nil, nil // 非TransferChecked指令，跳过
	}
}

// HeliusParseMarketBuyTx 通过 Helius Enhanced Transactions API 解析主网购买 token 的交易。
// heliusAPIKey 由调用方传入（取自 pos_global.HeliusApi.APIKey）。
func HeliusParseMarketBuyTx(heliusAPIKey string, txSig solana.Signature) (*entity.DecodedSolanaTransaction, error) {
	tokenMint := solana.MPK("ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH")
	const maxRetries = 5
	url := fmt.Sprintf("https://api-mainnet.helius-rpc.com/v0/transactions/?api-key=%s", heliusAPIKey)
	reqBody, err := json.Marshal(map[string]interface{}{"transactions": []string{txSig.String()}})
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	var rawTxList []heliusEnhancedTx
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := httpClient.Post(url, "application/json", bytes.NewReader(reqBody))
		if err != nil {
			if attempt == maxRetries {
				return nil, fmt.Errorf("Helius API 网络错误: %w", err)
			}
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			if attempt == maxRetries {
				return nil, fmt.Errorf("Helius API 限流，已重试 %d 次", maxRetries)
			}
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("Helius API 错误 %d: %s", resp.StatusCode, string(body))
		}

		if err := json.NewDecoder(resp.Body).Decode(&rawTxList); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("解析 Helius 响应失败: %w", err)
		}
		resp.Body.Close()
		break
	}

	if len(rawTxList) == 0 {
		return nil, fmt.Errorf("Helius 返回空数据")
	}
	htx := rawTxList[0]
	if htx.TransactionError != nil {
		return nil, fmt.Errorf("交易执行失败: %v", htx.TransactionError)
	}
	mintStr := tokenMint.String()

	// FromTokenAccount / FromNativeAccount：与原版保持一致，直接硬编码资金池账户
	fromTokenPubkey := solana.MPK("GjkvqFpZ5gqbzYEUAGsn5ozmFgM52JJDgso426DiLXbQ")
	fromNativePubkey := solana.MPK("GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL") // Raydium AMM

	// ToTokenAccount / ToNativeAccount：目标 mint 余额增加最多的账户
	// 买单中池子 token 余额是减少的（负数），自然不会被选中，无需额外排除
	var toTokenAcct, toUserAcct string
	var toDecimals uint8
	var maxIncrease int64

	for _, acct := range htx.AccountData {
		for _, c := range acct.TokenBalanceChanges {
			if c.Mint != mintStr {
				continue
			}
			var amt int64
			fmt.Sscanf(c.RawTokenAmount.TokenAmount, "%d", &amt)
			if amt > maxIncrease {
				maxIncrease = amt
				toTokenAcct = c.TokenAccount
				toUserAcct = c.UserAccount
				toDecimals = c.RawTokenAmount.Decimals
			}
		}
	}

	if toTokenAcct == "" {
		return nil, fmt.Errorf("未找到目标 mint token 余额增加的账户")
	}

	// 解析公钥
	feePayer, err := solana.PublicKeyFromBase58(htx.FeePayer)
	if err != nil {
		return nil, fmt.Errorf("解析 feePayer 失败: %w", err)
	}
	toTokenAccount, err := solana.PublicKeyFromBase58(toTokenAcct)
	if err != nil {
		return nil, fmt.Errorf("解析接收方 token 账户失败: %w", err)
	}
	toNativeAccount, err := solana.PublicKeyFromBase58(toUserAcct)
	if err != nil {
		return nil, fmt.Errorf("解析接收方 native 账户失败: %w", err)
	}
	mintPubkey := tokenMint

	// 构建账户列表（按 accountData 顺序，去重）
	var accounts []solana.PublicKey
	seen := make(map[string]bool)
	for _, acct := range htx.AccountData {
		if seen[acct.Account] {
			continue
		}
		seen[acct.Account] = true
		pk, err := solana.PublicKeyFromBase58(acct.Account)
		if err == nil {
			accounts = append(accounts, pk)
		}
	}

	decodedTx := &entity.DecodedSolanaTransaction{
		TxID:                 txSig,
		FromNativeAccount:    feePayer,
		FromTokenAccount:     fromTokenPubkey,
		FeePayer:             feePayer,
		Accounts:             accounts,
		Signatures:           []solana.Signature{txSig},
		TransferInstructions: []entity.DecodedSolTransferInst{},
		TransferCheckedInstructions: []entity.DecodedSolTransferCheckedInst{
			{
				FromTokenAccount:   fromTokenPubkey,
				FromNativeAccount:  fromNativePubkey,
				ToTokenAccount:     toTokenAccount,
				ToNativeAccount:    toNativeAccount,
				OwnerNativeAccount: fromNativePubkey,
				TokenMintAccount:   mintPubkey,
				Amount:             uint64(maxIncrease),
				Decimals:           toDecimals,
			},
		},
	}
	return decodedTx, nil
}
