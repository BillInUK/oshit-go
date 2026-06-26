package utils

import (
	"github.com/form3tech-oss/jwt-go"
	"github.com/oklog/ulid/v2"
	"github.com/pkg/errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// TokenMetadata struct to describe metadata in JWT.
type TokenMetadata struct {
	UserID      ulid.ULID
	Credentials map[string]interface{}
	Expires     int64
}

// ExtractTokenMetadata func to extract metadata from JWT.
func ExtractTokenMetadata(c *fiber.Ctx) (*TokenMetadata, error) {
	token, err := VerifyToken(c)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 1. 安全获取 "id"
	idVal, exists := claims["id"]
	if !exists {
		return nil, errors.New("missing 'id' in token claims")
	}
	idStr, ok := idVal.(string)
	if !ok {
		return nil, errors.New("invalid 'id' format")
	}
	userID, err := ulid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	// 2. 安全获取 "expires"
	expiresVal, exists := claims["expires"]
	if !exists {
		return nil, errors.New("missing 'expires' in token claims")
	}
	expiresFloat, ok := expiresVal.(float64)
	if !ok {
		return nil, errors.New("invalid 'expires' format")
	}
	expires := int64(expiresFloat)

	// 3. 安全获取 "credentials"
	credentialsVal, exists := claims["credentials"]
	if !exists {
		return nil, errors.New("missing 'credentials' in token claims")
	}
	credentials, ok := credentialsVal.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid 'credentials' format")
	}

	return &TokenMetadata{
		UserID:      userID,
		Credentials: credentials,
		Expires:     expires,
	}, nil
}

func ExtractToken(c *fiber.Ctx) string {
	bearToken := c.Get("Authorization")

	// Normally Authorization HTTP header.
	onlyToken := strings.Split(bearToken, " ")
	if len(onlyToken) == 2 {
		return onlyToken[1]
	}

	return ""
}

func VerifyToken(c *fiber.Ctx) (*jwt.Token, error) {
	tokenString := ExtractToken(c)

	token, err := jwt.Parse(tokenString, JwtKeyFunc)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func JwtKeyFunc(token *jwt.Token) (interface{}, error) {
	return JWTSecretBytes()
}
