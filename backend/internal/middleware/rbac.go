package middleware

import (
	"net/http"

	"github.com/kento/driver/backend/pkg/apperror"
)

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context())
			if claims == nil {
				apperror.WriteError(w, apperror.ErrUnauthorized)
				return
			}
			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			apperror.WriteError(w, apperror.ErrForbidden)
		})
	}
}
