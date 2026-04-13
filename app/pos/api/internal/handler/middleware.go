package handler

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"oshit-go/common/pkg/response"
	"oshit-go/common/utils"
	"strings"
)

// JWTAuthMiddleware 直接解析 JWT 签名（无状态，不依赖 dtoken storage），
// 将 loginID（nativeAccount）存入 c.Locals("nativeAccount")。
func JWTAuthMiddleware(c *fiber.Ctx) error {
	bearerToken := c.Get("Authorization")
	tokenValue := strings.TrimPrefix(bearerToken, "Bearer ")
	if tokenValue == "" {
		return response.UnAuthorizedError(c, "unauthorized")
	}

	loginID, err := parseLoginIDFromJWT(tokenValue)
	if err != nil || loginID == "" {
		return response.UnAuthorizedError(c, "unauthorized")
	}

	if _, err := solana.PublicKeyFromBase58(loginID); err != nil {
		return response.UnAuthorizedError(c, "unauthorized")
	}

	c.Locals("nativeAccount", loginID)
	return c.Next()
}

func parseLoginIDFromJWT(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(utils.JWT_SECRET_KEY), nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	loginID, ok := claims["loginId"].(string)
	if !ok {
		return "", fmt.Errorf("loginId not found")
	}
	return loginID, nil
}
