package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"

	"oshit-go/common/pkg/entity"
)

// TransferCheckedParams 定义返回的TransferChecked指令参数结构体
type TransferCheckedParams struct {
	SourceAccount string
	DestAccount   string
	MintAccount   string
	OwnerAccount  string
	Amount        uint64
	Decimals      uint8
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

// ParseMarketBuyTx 优化版核心函数
//
// 说明:
//   - 该函数仅限主网使用，专门用来解析主网购买token的交易
//   - 修改为通用版本，不再硬编码特定市场账户
func ParseMarketBuyTx(rpcClient *rpc.Client, txSig solana.Signature) (*entity.DecodedSolanaTransaction, error) {
	sourceAccount := solana.MPK("GjkvqFpZ5gqbzYEUAGsn5ozmFgM52JJDgso426DiLXbQ")
	tokenMint := solana.MPK("ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH")

	txResp, err := getTransaction(rpcClient, txSig)
	if err != nil {
		return nil, fmt.Errorf("获取交易失败: %w", err)
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txResp.Transaction.GetBinary()))
	if err != nil {
		return nil, fmt.Errorf("解析交易失败: %w", err)
	}

	fullAccountList := buildFullAccountList(tx, txResp)

	transferCheckedInst, err := parseInnerInstructionsForTransferChecked(txResp, fullAccountList, sourceAccount, tokenMint)
	if err != nil {
		return nil, fmt.Errorf("解析Inner Instructions失败: %w", err)
	}
	if transferCheckedInst != nil {
		decodedTx := &entity.DecodedSolanaTransaction{
			TxID:                        txSig,
			FromNativeAccount:           tx.Message.AccountKeys[0],
			FromTokenAccount:            solana.MustPublicKeyFromBase58(transferCheckedInst.SourceAccount),
			FeePayer:                    tx.Message.AccountKeys[0],
			RefBlockHash:                tx.Message.RecentBlockhash,
			Accounts:                    fullAccountList,
			Signatures:                  tx.Signatures,
			TransferInstructions:        []entity.DecodedSolTransferInst{},
			TransferCheckedInstructions: []entity.DecodedSolTransferCheckedInst{},
			ComputeUnitPrice:            0,
			ComputeUnitLimit:            0,
			EstimateFee:                 0,
		}

		transferChecked := entity.DecodedSolTransferCheckedInst{
			FromTokenAccount:   solana.MustPublicKeyFromBase58(transferCheckedInst.SourceAccount),
			FromNativeAccount:  solana.MustPublicKeyFromBase58(transferCheckedInst.OwnerAccount),
			ToTokenAccount:     solana.MustPublicKeyFromBase58(transferCheckedInst.DestAccount),
			ToNativeAccount:    solana.PublicKey{},
			OwnerNativeAccount: solana.MustPublicKeyFromBase58(transferCheckedInst.OwnerAccount),
			TokenMintAccount:   solana.MustPublicKeyFromBase58(transferCheckedInst.MintAccount),
			Amount:             transferCheckedInst.Amount,
			Decimals:           transferCheckedInst.Decimals,
		}
		decodedTx.TransferCheckedInstructions = append(decodedTx.TransferCheckedInstructions, transferChecked)

		destAccount := solana.MustPublicKeyFromBase58(transferCheckedInst.DestAccount)

		isSystemAccount := destAccount.Equals(solana.SystemProgramID) ||
			destAccount.Equals(solana.TokenProgramID) ||
			destAccount.Equals(solana.MPK("GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL")) ||
			destAccount.Equals(solana.MPK("8AgyxWiUW4Wczmqsh89vz6cjkmkPvJbEeU7WS51jkGWv"))

		if !isSystemAccount {
			if owner, err := GetNativeAccountByTokenAccount(rpcClient, destAccount); err == nil && owner != nil {
				decodedTx.TransferCheckedInstructions[0].ToNativeAccount = *owner
			} else {
				decodedTx.TransferCheckedInstructions[0].ToNativeAccount = destAccount
			}
		}

		return decodedTx, nil
	}

	return nil, fmt.Errorf("未找到符合条件的TransferChecked指令")
}

func buildFullAccountList(tx *solana.Transaction, txResp *rpc.GetTransactionResult) []solana.PublicKey {
	fullList := make([]solana.PublicKey, len(tx.Message.AccountKeys))
	copy(fullList, tx.Message.AccountKeys)

	if txResp.Meta != nil {
		for _, addr := range txResp.Meta.LoadedAddresses.ReadOnly {
			fullList = append(fullList, addr)
		}
		for _, addr := range txResp.Meta.LoadedAddresses.Writable {
			fullList = append(fullList, addr)
		}
	}
	return fullList
}

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

func parseInnerInstructionsForTransferChecked(txResp *rpc.GetTransactionResult, fullAccountList []solana.PublicKey, source, mint solana.PublicKey) (*TransferCheckedParams, error) {
	if txResp.Meta == nil {
		return nil, nil
	}

	extractAmountFromBalances := func() (uint64, uint8, solana.PublicKey, bool) {
		if len(txResp.Meta.PreTokenBalances) == 0 || len(txResp.Meta.PostTokenBalances) == 0 {
			fmt.Printf("调试: 无token balances数据\n")
			return 0, 0, solana.PublicKey{}, false
		}

		var marketPreBalance, marketPostBalance uint64
		var decimals uint8
		var userAccount solana.PublicKey

		mintStr := mint.String()
		marketAccount := solana.MPK("GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL")
		marketStr := marketAccount.String()

		fmt.Printf("调试: 查找市场账户=%s, mint=%s的余额变化\n", marketStr, mintStr)
		fmt.Printf("调试: PreTokenBalances数量=%d, PostTokenBalances数量=%d\n",
			len(txResp.Meta.PreTokenBalances), len(txResp.Meta.PostTokenBalances))

		for _, pre := range txResp.Meta.PreTokenBalances {
			fmt.Printf("调试: Pre - Owner:%s, Mint:%s, Amount:%s, Decimals:%d\n",
				pre.Owner, pre.Mint, pre.UiTokenAmount.Amount, pre.UiTokenAmount.Decimals)
			if pre.Mint.String() == mintStr && pre.Owner.String() == marketStr {
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
					var amount uint64
					fmt.Sscanf(post.UiTokenAmount.Amount, "%d", &amount)
					marketPostBalance = amount
					fmt.Printf("调试: 找到市场 Post余额: %d\n", amount)
				} else if post.Owner.String() != marketStr {
					userAccount = *post.Owner
					fmt.Printf("调试: 找到用户账户: %s, 余额: %s\n", userAccount, post.UiTokenAmount.Amount)
				}
			}
		}

		fmt.Printf("调试: marketPreBalance=%d, marketPostBalance=%d, userAccount=%s\n",
			marketPreBalance, marketPostBalance, userAccount)

		if marketPreBalance > marketPostBalance && !userAccount.IsZero() {
			amount := marketPreBalance - marketPostBalance
			fmt.Printf("调试: 计算金额: %d - %d = %d\n", marketPreBalance, marketPostBalance, amount)
			return amount, decimals, userAccount, true
		}

		fmt.Printf("调试: 无法提取金额\n")
		return 0, 0, solana.PublicKey{}, false
	}

	if len(txResp.Meta.InnerInstructions) == 0 {
		if amount, decimals, destAccount, ok := extractAmountFromBalances(); ok {
			return &TransferCheckedParams{
				SourceAccount: source.String(),
				DestAccount:   destAccount.String(),
				MintAccount:   mint.String(),
				OwnerAccount:  "GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL",
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
			var progID solana.PublicKey
			if int(innerInst.ProgramIDIndex) >= len(fullAccountList) {
				continue
			} else {
				progID = fullAccountList[innerInst.ProgramIDIndex]
			}

			isTokenProgram := progID.Equals(tokenProgramID)
			isJupiterWrapper := progID.Equals(jupiterWrapperProgramID)
			isTransitSwap := progID.Equals(transitSwapProgramID)

			if !isTokenProgram && !isJupiterWrapper && !isTransitSwap {
				continue
			}

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
				continue
			}

			if isTokenProgram {
				params, err := decodeTokenInstructionForTransferChecked(innerInst, fullAccountList, source, mint)
				if err != nil {
					fmt.Printf("解码Token指令失败: %v\n", err)
					continue
				}
				if params != nil {
					return params, nil
				}
			}

			if isJupiterWrapper || isTransitSwap {
				params, err := decodeTokenInstructionForTransferChecked(innerInst, fullAccountList, source, mint)
				if err == nil && params != nil {
					return params, nil
				}

				var tokenProgramIndex uint16
				if isJupiterWrapper && len(innerInst.Accounts) >= 1 {
					tokenProgramIndex = innerInst.Accounts[0]
				} else if isTransitSwap && len(innerInst.Accounts) >= 7 {
					tokenProgramIndex = innerInst.Accounts[6]
				} else {
					continue
				}

				if int(tokenProgramIndex) < len(fullAccountList) {
					tokenProgram := fullAccountList[tokenProgramIndex]
					if tokenProgram.Equals(tokenProgramID) {
						var destAccount solana.PublicKey
						var sourceAccount solana.PublicKey

						if isJupiterWrapper && len(innerInst.Accounts) >= 4 {
							destIndex := innerInst.Accounts[2]
							sourceIndex := innerInst.Accounts[3]
							if int(destIndex) < len(fullAccountList) && int(sourceIndex) < len(fullAccountList) {
								destAccount = fullAccountList[destIndex]
								sourceAccount = fullAccountList[sourceIndex]
							}
						} else if isTransitSwap && len(innerInst.Accounts) >= 13 {
							destIndex := innerInst.Accounts[5]
							sourceIndex := innerInst.Accounts[10]
							if int(destIndex) < len(fullAccountList) && int(sourceIndex) < len(fullAccountList) {
								destAccount = fullAccountList[destIndex]
								sourceAccount = fullAccountList[sourceIndex]
							}
						}

						if !destAccount.IsZero() && !sourceAccount.IsZero() {
							if !sourceAccount.Equals(source) {
								continue
							}

							amount, decimals, balanceDestAccount, ok := extractAmountFromBalances()
							if !ok {
								amount = 0
								decimals = 0
							} else if !balanceDestAccount.IsZero() && !balanceDestAccount.Equals(destAccount) {
								destAccount = balanceDestAccount
							}

							ownerAccount := "GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL"
							if isTransitSwap {
								ownerAccount = "8AgyxWiUW4Wczmqsh89vz6cjkmkPvJbEeU7WS51jkGWv"
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

	return nil, nil
}

func cleanJSONQuotes(data []byte) []byte {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		return data[1 : len(data)-1]
	}
	return data
}

func isTokenInstruction(opcode byte) bool {
	return opcode == 13
}

func decodeTokenInstructionForTransferChecked(inst rpc.CompiledInstruction, fullAccountList []solana.PublicKey, source, mint solana.PublicKey) (*TransferCheckedParams, error) {
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

	dataJSON, err := inst.Data.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("解析Data失败: %w", err)
	}
	cleanData := cleanJSONQuotes(dataJSON)
	rawData, err := base58.Decode(string(cleanData))
	if err != nil {
		return nil, fmt.Errorf("base58解码失败: %v", err)
	}

	if len(rawData) >= 1 {
		opcode := rawData[0]
		fmt.Printf("调试: 指令opcode=%d (0x%x), 数据长度=%d\n", opcode, opcode, len(rawData))

		if opcode == 13 && len(rawData) >= 10 {
			var amount uint64
			for i := 0; i < 8; i++ {
				amount |= uint64(rawData[1+i]) << (8 * i)
			}
			decimals := rawData[9]

			fmt.Printf("调试: 找到TransferChecked指令, amount=%d, decimals=%d\n", amount, decimals)

			if len(accounts) >= 4 {
				sourceAccount := accounts[0].PublicKey
				mintAccount := accounts[1].PublicKey
				destAccount := accounts[2].PublicKey
				ownerAccount := accounts[3].PublicKey

				fmt.Printf("调试: 账户: source=%s, mint=%s, dest=%s, owner=%s\n",
					sourceAccount, mintAccount, destAccount, ownerAccount)

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

	decodedInst, err := token.DecodeInstruction(accounts, rawData)
	if err != nil {
		return nil, fmt.Errorf("解码Token指令失败: %w", err)
	}

	switch decoded := decodedInst.Impl.(type) {
	case *token.TransferChecked:
		sourceAccount := decoded.GetSourceAccount().PublicKey
		if !sourceAccount.Equals(source) {
			return nil, nil
		}
		mintAccount := decoded.GetMintAccount().PublicKey
		if !mintAccount.Equals(mint) {
			return nil, nil
		}

		return &TransferCheckedParams{
			SourceAccount: sourceAccount.String(),
			DestAccount:   decoded.GetDestinationAccount().PublicKey.String(),
			MintAccount:   decoded.GetMintAccount().PublicKey.String(),
			OwnerAccount:  decoded.GetOwnerAccount().PublicKey.String(),
			Amount:        *decoded.Amount,
			Decimals:      *decoded.Decimals,
		}, nil
	default:
		return nil, nil
	}
}

// HeliusParseMarketBuyTx 通过 Helius Enhanced Transactions API 解析主网购买 token 的交易。
func HeliusParseMarketBuyTx(heliusAPIKey string, txSig solana.Signature) (*entity.DecodedSolanaTransaction, error) {
	tokenMint := solana.MPK("ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH")
	const maxRetries = 5
	url := fmt.Sprintf("https://mainnet.helius-rpc.com/v0/transactions/?api-key=%s", heliusAPIKey)
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

	fromTokenPubkey := solana.MPK("GjkvqFpZ5gqbzYEUAGsn5ozmFgM52JJDgso426DiLXbQ")
	fromNativePubkey := solana.MPK("GpMZbSM2GgvTKHJirzeGfMFoaZ8UR2X7F4v8vHTvxFbL")

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
