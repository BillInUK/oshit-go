package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

const (
	defaultConfigKeyEncFile  = "/etc/credentials/oshit-config-key.enc"
	defaultPrivKeyEncFile    = "/etc/credentials/oshit-priv-key.enc"
	defaultAWSRegion         = "ap-southeast-1"
)

// LoadConfigDecryptKey loads the AES decryption key for configuration fields via KMS envelope decryption.
//
// Reads base64-encoded KMS ciphertext from $CONFIG_KEY_ENC_FILE
// (default: /etc/credentials/oshit-config-key.enc), calls KMS Decrypt,
// and returns the plaintext AES key.
//
// Local dev: set CONFIG_KEY_ENC_FILE=./etc/dev-config-key.enc
// and AWS_PROFILE=oshit-testnet (or whichever profile has kms:Decrypt).
func LoadConfigDecryptKey() (string, error) {
	encFile := os.Getenv("CONFIG_KEY_ENC_FILE")
	if encFile == "" {
		encFile = defaultConfigKeyEncFile
	}
	return loadKMSDecryptKey(encFile, "config")
}

// LoadPrivKeyDecryptKey loads the AES decryption key for Solana private keys via KMS envelope decryption.
//
// Reads base64-encoded KMS ciphertext from $PRIV_KEY_ENC_FILE
// (default: /etc/credentials/oshit-priv-key.enc), calls KMS Decrypt,
// and returns the plaintext AES key used to decrypt encrypted_key in t_service_key / Nacos.
//
// Local dev: set PRIV_KEY_ENC_FILE=./etc/dev-priv-key.enc
func LoadPrivKeyDecryptKey() (string, error) {
	encFile := os.Getenv("PRIV_KEY_ENC_FILE")
	if encFile == "" {
		encFile = defaultPrivKeyEncFile
	}
	return loadKMSDecryptKey(encFile, "priv-key")
}

// loadKMSDecryptKey reads a base64-encoded KMS ciphertext from the given file,
// calls AWS KMS Decrypt, and returns the plaintext key.
func loadKMSDecryptKey(encFile, label string) (string, error) {
	ciphertextB64, err := os.ReadFile(encFile)
	if err != nil {
		return "", fmt.Errorf("read encrypted key file %s: %w", encFile, err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(ciphertextB64)))
	if err != nil {
		return "", fmt.Errorf("base64 decode encrypted key (%s): %w", label, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = defaultAWSRegion
	}
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("load AWS config: %w", err)
	}

	client := kms.NewFromConfig(awsCfg)
	result, err := client.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: ciphertext,
	})
	if err != nil {
		return "", fmt.Errorf("KMS Decrypt (%s): %w", label, err)
	}

	key := strings.TrimSpace(string(result.Plaintext))
	if key == "" {
		return "", fmt.Errorf("KMS decrypted an empty key (%s)", label)
	}

	fmt.Printf("[credential] AES key loaded via KMS Decrypt (%s)\n", label)
	return key, nil
}
