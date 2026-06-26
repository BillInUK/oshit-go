package test

import (
	"os"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
)

func testRPCURL() string {
	if v := strings.TrimSpace(os.Getenv("OSHIT_TEST_RPC_URL")); v != "" {
		return v
	}
	return "https://api.devnet.solana.com"
}

func loadTestPrivateKey(t *testing.T, name string) solana.PrivateKey {
	t.Helper()

	envName := "OSHIT_TEST_" + strings.ToUpper(name) + "_KEYPAIR"
	path := strings.TrimSpace(os.Getenv(envName))
	if path == "" {
		t.Skipf("%s is required for this integration test", envName)
	}

	privateKey, err := solana.PrivateKeyFromSolanaKeygenFile(path)
	if err != nil {
		t.Fatalf("load %s: %v", envName, err)
	}
	return privateKey
}
