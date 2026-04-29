package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/AgataPalma/biblios/internal/tokenstore"
	"github.com/golang-jwt/jwt/v5"
)

func Authenticate(jwtSecret string, store *tokenstore.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims := apictx.Claims{}
			token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
				return []byte(jwtSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}
			exists, err := store.SessionExists(r.Context(), claims.UserID, claims.ID)
			if err != nil || !exists {
				http.Error(w, `{"error":"session not found or expired"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), apictx.UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
