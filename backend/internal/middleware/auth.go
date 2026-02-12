package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/kento/driver/backend/pkg/apperror"
	"github.com/kento/driver/backend/pkg/jwt"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				apperror.WriteError(w, apperror.ErrUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				apperror.WriteError(w, apperror.ErrUnauthorized)
				return
			}

			claims, err := jwt.Parse(parts[1], secret)
			if err != nil {
				apperror.WriteError(w, apperror.ErrUnauthorized)
				return
			}

			if claims.TokenType != "access" {
				apperror.WriteError(w, apperror.ErrUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(ctx context.Context) *jwt.Claims {
	claims, ok := ctx.Value(ClaimsKey).(*jwt.Claims)
	if !ok {
		return nil
	}
	return claims
}
