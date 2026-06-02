package utils

import (
	"github.com/gagliardetto/solana-go"
	"testing"
)

func TestGenRandomAddress(t *testing.T) {
	wallet := solana.NewWallet()
	t.Logf("create wallet %v\n", wallet.PublicKey())
}
