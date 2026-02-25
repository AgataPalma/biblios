package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/AgataPalma/biblios/internal/apictx"
	"github.com/golang-jwt/jwt/v5"
)

func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var authHeader string = r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			var tokenString string = strings.TrimPrefix(authHeader, "Bearer ")
			var claims apictx.Claims = apictx.Claims{}
			var err error
			var token *jwt.Token

			token, err = jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			var ctx context.Context = context.WithValue(r.Context(), apictx.UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
