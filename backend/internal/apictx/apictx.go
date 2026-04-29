package apictx

import "github.com/golang-jwt/jwt/v5"

type contextKey string

const UserClaimsKey contextKey = "userClaims"

type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}
