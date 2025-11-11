package middleware

import (
	"errors"
	"net/http"
	"strings"

	"internal-work-management-service/internal/app/auth"
	"internal-work-management-service/internal/app/handlers"
)

var errUnauthorized = errors.New("unauthorized")

// JWT verifies Authorization: Bearer tokens using the provided secret.
func JWT(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
			if token == "" {
				handlers.WriteError(w, http.StatusUnauthorized, errUnauthorized)
				return
			}

			claims, err := auth.Parse(token, secret)
			if err != nil {
				handlers.WriteError(w, http.StatusUnauthorized, errUnauthorized)
				return
			}

			ctx := auth.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
