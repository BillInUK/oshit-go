package entity

import "github.com/gagliardetto/solana-go"

type DecodedSolTransferInst struct {
	FromNativeAccount solana.PublicKey
	ToNativeAccount   solana.PublicKey
	Amount            uint64
}

type DecodedServiceTransferInst struct {
	FromNativeAccount string
	ToNativeAccount   string
	Amount            float64
}

type DecodedSolTransferCheckedInst struct {
	FromTokenAccount   solana.PublicKey
	FromNativeAccount  solana.PublicKey
	ToTokenAccount     solana.PublicKey
	ToNativeAccount    solana.PublicKey
	OwnerNativeAccount solana.PublicKey
	TokenMintAccount   solana.PublicKey
	Amount             uint64
	Decimals           uint8
}

type DecodedServiceTransferCheckedInst struct {
	FromTokenAccount   string
	FromNativeAccount  string
	ToTokenAccount     string
	ToNativeAccount    string
	OwnerNativeAccount string
	TokenMintAccount   string
	Amount             float64
	Decimals           int32
}

type DecodedSolOfficeTransferTx struct {
	TransferInst DecodedSolTransferCheckedInst
	RewardInst   DecodedSolTransferCheckedInst
	DexInst      DecodedSolTransferInst
	TxId         solana.Signature
}

type DecodedSolanaTransaction struct {
	TxID                        solana.Signature
	FromNativeAccount           solana.PublicKey
	FromTokenAccount            solana.PublicKey
	FeePayer                    solana.PublicKey
	RefBlockHash                solana.Hash
	Accounts                    []solana.PublicKey
	Signatures                  []solana.Signature
	TransferInstructions        []DecodedSolTransferInst
	TransferCheckedInstructions []DecodedSolTransferCheckedInst
	ComputeUnitPrice            uint64
	ComputeUnitLimit            uint64
	EstimateFee                 uint64
}

type DecodedServiceTransaction struct {
	TxID              string
	RefBlockHash      string
	FromNativeAccount string
	FromTokenAccount  string
	FeePayer          string
	Accounts          []string
	Signatures        []string
	ToDexInst         DecodedServiceTransferInst
	TransferTokenInst DecodedServiceTransferCheckedInst
	RewardInst        DecodedServiceTransferCheckedInst
	RewardInviterInst []DecodedServiceTransferCheckedInst
	EstimateFee       float64
}

type DecodedSwapTransaction struct {
	TxID                          string
	TxType                        int
	RefBlockHash                  string
	FromNativeAccount             string
	FromTokenAccount              string
	FeePayer                      string
	Accounts                      []string
	Signatures                    []string
	InputTransferSwapInst         DecodedServiceTransferInst
	InputTransferCheckedSwapInst  DecodedServiceTransferCheckedInst
	OutputTransferCheckedSwapInst DecodedServiceTransferCheckedInst
}

type SOLRpcRequestBody struct {
	JsonRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type SOLRpcGetBlockResponse struct {
	JsonRPC string `json:"jsonrpc"`
	Result  struct {
		BlockHeight       int64  `json:"blockHeight"`
		BlockTime         int64  `json:"blockTime"`
		Blockhash         string `json:"blockhash"`
		ParentSlot        int64  `json:"parentSlot"`
		PreviousBlockhash string `json:"previousBlockhash"`
		Transactions      []struct {
			Meta struct {
				ComputeUnitsConsumed int         `json:"computeUnitsConsumed"`
				Err                  interface{} `json:"err"`
				Fee                  int         `json:"fee"`
				LogMessages          []string    `json:"logMessages"`
			} `json:"meta"`
			Transaction struct {
				Message struct {
					AccountKeys []struct {
						Pubkey   string `json:"pubkey"`
						Signer   bool   `json:"signer"`
						Writable bool   `json:"writable"`
					} `json:"accountKeys"`
					Instructions []struct {
						ProgramID string `json:"programId"`
						Data      string `json:"data"`
					} `json:"instructions"`
					RecentBlockhash string `json:"recentBlockhash"`
				} `json:"message"`
				Signatures []string `json:"signatures"`
			} `json:"transaction"`
		} `json:"transactions"`
	} `json:"result"`
	ID int `json:"id"`
}
