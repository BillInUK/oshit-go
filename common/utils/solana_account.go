package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
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

func ParseDbArray(input string) []string {
	if len(input) < 2 {
		return []string{}
	}

	input = strings.TrimSpace(input)
	if input[0] == '{' && input[len(input)-1] == '}' {
		input = input[1 : len(input)-1]
	} else {
		return []string{}
	}

	if input == "" {
		return []string{}
	}

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
	accountInfo, err := rpcClient.GetAccountInfo(context.Background(), tokenAccountPubKey)
	if err != nil {
		fmt.Printf("Error retrieving account info: %v\n", err)
		return nil, err
	}
	if accountInfo.Value == nil || accountInfo.Value.Data.GetBinary() == nil {
		fmt.Println("Account data is not available")
		return nil, err
	}
	data := accountInfo.Value.Data.GetBinary()
	if len(data) < 64 {
		return nil, errors.New("account data is too short to contain required information")
	}
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

// GetTransactionCountForReward 查看地址的交易数是否超过count
func GetTransactionCountForReward(rpcClient *rpc.Client, nativeAccount solana.PublicKey) (int, error) {
	var limit = 1000
	nativeTransactions, err := rpcClient.GetSignaturesForAddressWithOpts(context.Background(), nativeAccount, &rpc.GetSignaturesForAddressOpts{
		Limit:      &limit,
		Commitment: rpc.CommitmentFinalized,
	})
	if err != nil {
		return 0, err
	}
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
