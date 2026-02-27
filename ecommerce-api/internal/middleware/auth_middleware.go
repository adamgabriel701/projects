package middleware

import (
	"context"
	"net/http"
	"strings"
	"ecommerce-api/internal/domain"
	"ecommerce-api/internal/service"
)

type contextKey string

const ContextUserID contextKey = "user_id"
const ContextUserRole contextKey = "user_role"

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "Token não fornecido"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error": "Formato de token inválido"}`, http.StatusUnauthorized)
				return
			}

			userID, role, err := authService.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, `{"error": "Token inválido"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, userID)
			ctx = context.WithValue(ctx, ContextUserRole, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) string {
	userID, _ := r.Context().Value(ContextUserID).(string)
	return userID
}

func GetUserRole(r *http.Request) domain.UserRole {
	role, _ := r.Context().Value(ContextUserRole).(domain.UserRole)
	return role
}
