package utils

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	"testing"
)

func TestGenRandomAddress(t *testing.T) {
	wallet := solana.NewWallet()
	t.Logf("create wallet %v\n", wallet.PublicKey())
}

func TestGetATA(t *testing.T) {
	nativeAccount := solana.MPK("AUq3iXbJjBjED6ZN2JqJDdZ3mknd7cXv9HQqNdYuEZ6q")
	mint := solana.MPK("wtnrTujJqBRUknLRhQQcUSwzAzx8LvcxKXEuBwvFnJM")
	tokenAccount, _, err := solana.FindAssociatedTokenAddress(nativeAccount, mint)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", tokenAccount)
}
