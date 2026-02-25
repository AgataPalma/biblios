package auth

import (
	"fmt"
	"time"

	"crypto/rand"
	"encoding/hex"
	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/golang-jwt/jwt/v5"
)

func generateTokenID() (string, error) {
	var bytes []byte = make([]byte, 16)
	var err error

	_, err = rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate token ID: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

func GenerateToken(userID string, isAdmin bool, secret string) (string, error) {
	var tokenID string
	var err error

	tokenID, err = generateTokenID()
	if err != nil {
		return "", err
	}

	var claims apictx.Claims = apictx.Claims{
		UserID:  userID,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	var token *jwt.Token = jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	var signed string
	signed, err = token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}
