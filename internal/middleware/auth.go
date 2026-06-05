package middleware

import (
	"context"
	"net/http"
	"strings"

	"goshen/backend/pkg/response"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const AdminIDKey contextKey = "admin_id"

func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				response.Unauthorized(w, "missing or invalid token")
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				response.Unauthorized(w, "invalid token")
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.Unauthorized(w, "invalid token claims")
				return
			}
			ctx := context.WithValue(r.Context(), AdminIDKey, claims["sub"])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
