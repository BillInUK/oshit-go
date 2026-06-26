package utils

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/form3tech-oss/jwt-go"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT = 5 * 24 * 60 // 5*24*60
	JWT_REFRESH_KEY                     = "refresh"
	JWT_REFRESH_KEY_EXPIRE_HOURS_COUNT  = 5
)

var (
	jwtSecretMu sync.RWMutex
	jwtSecret   string
)

func SetJWTSecret(secret string) error {
	if strings.TrimSpace(secret) == "" {
		return fmt.Errorf("jwt secret is empty")
	}
	jwtSecretMu.Lock()
	defer jwtSecretMu.Unlock()
	jwtSecret = secret
	return nil
}

func JWTSecret() string {
	jwtSecretMu.RLock()
	defer jwtSecretMu.RUnlock()
	return jwtSecret
}

func JWTSecretBytes() ([]byte, error) {
	secret := JWTSecret()
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("jwt secret is not initialized")
	}
	return []byte(secret), nil
}

func LoadJWTSecretFromEnv() error {
	if strings.TrimSpace(JWTSecret()) != "" {
		return nil
	}
	secret := strings.TrimSpace(os.Getenv("OSHIT_JWT_SECRET"))
	if secret == "" {
		return nil
	}
	return SetJWTSecret(secret)
}

// Tokens struct to describe tokens object.
type Tokens struct {
	Access  string
	Refresh string
}

// GenerateNewTokens func for generate a new Access & Refresh tokens.
func GenerateNewTokens(id string, credentials map[string]interface{}) (*Tokens, error) {
	// Generate JWT Access token.
	accessToken, err := generateNewAccessToken(id, credentials)
	if err != nil {
		// Return token generation error.
		return nil, err
	}

	// Generate JWT Refresh token.
	refreshToken, err := generateNewRefreshToken()
	if err != nil {
		// Return token generation error.
		return nil, err
	}

	return &Tokens{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

// GenerateNewOAuthTokens func for generate a new Access & Refresh tokens.
func GenerateNewOAuthTokens(id string, credentials map[string]interface{}) (*Tokens, error) {
	// Generate JWT Access token.
	accessToken, err := generateNewAccessToken(id, credentials)
	if err != nil {
		// Return token generation error.
		return nil, err
	}

	// Generate JWT Refresh token.
	refreshToken, err := generateNewRefreshToken()
	if err != nil {
		// Return token generation error.
		return nil, err
	}

	return &Tokens{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func generateNewAccessToken(id string, credentials map[string]interface{}) (string, error) {
	// Create a new claims.
	claims := jwt.MapClaims{}

	// Set public claims:
	claims["id"] = id
	claims["expires"] = time.Now().Add(time.Minute * JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT).Unix()
	claims["credentials"] = credentials

	// Create a new JWT access token with claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate token.
	secret, err := JWTSecretBytes()
	if err != nil {
		return "", err
	}
	t, err := token.SignedString(secret)
	if err != nil {
		// Return error, it JWT token generation failed.
		return "", err
	}

	return t, nil
}

func generateNewRefreshToken() (string, error) {
	// Create a new SHA256 hash.
	hash := sha256.New()

	// Create a new now date and time string with salt.
	refresh := JWT_REFRESH_KEY + time.Now().String()

	// See: https://pkg.go.dev/io#Writer.Write
	_, err := hash.Write([]byte(refresh))
	if err != nil {
		// Return error, it refresh token generation failed.
		return "", err
	}

	// Set expiration time.
	expireTime := fmt.Sprint(time.Now().Add(time.Hour * JWT_REFRESH_KEY_EXPIRE_HOURS_COUNT).Unix())

	// Create a new refresh token (sha256 string with salt + expire time).
	t := hex.EncodeToString(hash.Sum(nil)) + "." + expireTime

	return t, nil
}

// ParseRefreshToken func for parse second argument from refresh token.
func ParseRefreshToken(refreshToken string) (int64, error) {
	return strconv.ParseInt(strings.Split(refreshToken, ".")[1], 0, 64)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	randomBytes := make([]byte, length)
	for i := range randomBytes {
		randomBytes[i] = charset[r.Intn(len(charset))]
	}
	return string(randomBytes)
}

// ParseRsaPublicKeyFromPemStr 从PEM字符串解析RSA公钥
func ParseRsaPublicKeyFromPemStr(pubKeyPemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubKeyPemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key: %v", err)
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, fmt.Errorf("key type is not RSA")
	}
}
