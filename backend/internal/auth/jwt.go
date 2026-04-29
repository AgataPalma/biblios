package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/golang-jwt/jwt/v5"
)

func generateTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func GenerateToken(userID string, role apictx.Role, secret string) (string, string, error) {
	tokenID, err := generateTokenID()
	if err != nil {
		return "", "", err
	}
	claims := apictx.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", "", fmt.Errorf("sign token: %w", err)
	}
	return signed, tokenID, nil
}
