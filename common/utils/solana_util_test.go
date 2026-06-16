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
	mint := solana.MPK("ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH")
	tokenAccount, _, err := solana.FindAssociatedTokenAddress(nativeAccount, mint)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", tokenAccount)
}

func TestLoadPrivateKey(t *testing.T) {
	privateKey, err := solana.PrivateKeyFromBase58("5ggcAiHkMb5KE4qMyzbCVcWaAU48E97hAJjt5rwZZRGCZFMfXFASEXts3YuLqt7ub7PUhiDSwpKDmXH196Ei9t4T")
	if err != nil {
		panic(err)
	}
	publicKey := privateKey.PublicKey()
	fmt.Printf("public key: %v\n", publicKey)
}
