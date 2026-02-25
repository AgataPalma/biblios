package middleware

import (
	"net/http"

	"github.com/AgataPalma/biblios/internal/apictx"
)

func RequireRole(roles ...apictx.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var claims apictx.Claims
			var ok bool

			claims, ok = r.Context().Value(apictx.UserClaimsKey).(apictx.Claims)
			if !ok {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		})
	}
}
