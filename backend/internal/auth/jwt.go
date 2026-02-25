package auth

import (
	"fmt"
	"time"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID string, isAdmin bool, secret string) (string, error) {
	var claims apictx.Claims = apictx.Claims{
		UserID:  userID,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	var token *jwt.Token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	var signed string
	var err error
	signed, err = token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}
