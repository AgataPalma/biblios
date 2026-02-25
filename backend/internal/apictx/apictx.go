package apictx

import "github.com/golang-jwt/jwt/v5"

type contextKey string

const UserClaimsKey contextKey = "userClaims"

type Claims struct {
	UserID  string `json:"user_id"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}
