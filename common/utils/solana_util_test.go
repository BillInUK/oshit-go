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
	nativeAccount := solana.MPK("9eXCGjngWmLdFwcy2NiUUp1aAR7TcfTUV7G7TQURWn8a")
	mint := solana.MPK("ShitJuMfPKCQU7LedLERFYapDta7CCdKExPWX2gETRH")
	tokenAccount, _, err := solana.FindAssociatedTokenAddress(nativeAccount, mint)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v\n", tokenAccount)
}

func TestLoadPrivateKey(t *testing.T) {
	privateKey, err := solana.PrivateKeyFromBase58("5EEB7yAG8pyHLZDfk7zzzGfsMcKfgMCwFbr6pwoaXKDDVQa78umAk3m6ehr7em8ZsWubg7FnjWBvZHgarxHKFJk2")
	if err != nil {
		panic(err)
	}
	publicKey := privateKey.PublicKey()
	fmt.Printf("public key: %v\n", publicKey)
}
