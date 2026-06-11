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
	defaultEncKeyFile = "/etc/credentials/oshit-config-key.enc"
	defaultAWSRegion  = "ap-southeast-1"
)

// LoadConfigDecryptKey loads the AES decryption key via KMS envelope decryption.
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
		encFile = defaultEncKeyFile
	}

	ciphertextB64, err := os.ReadFile(encFile)
	if err != nil {
		return "", fmt.Errorf("read encrypted key file %s: %w (set CONFIG_KEY_ENC_FILE for local dev)", encFile, err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(ciphertextB64)))
	if err != nil {
		return "", fmt.Errorf("base64 decode encrypted key: %w", err)
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
		return "", fmt.Errorf("KMS Decrypt: %w", err)
	}

	key := strings.TrimSpace(string(result.Plaintext))
	if key == "" {
		return "", fmt.Errorf("KMS decrypted an empty key")
	}

	fmt.Printf("[credential] AES key loaded via KMS Decrypt\n")
	return key, nil
}
