package middleware

import (
	"context"
	"net/http"
	"strings"
	"task-manager/internal/service"
)

type contextKey string

const ContextUserID contextKey = "user_id"

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Token não fornecido", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "Formato de token inválido", http.StatusUnauthorized)
				return
			}

			userID, err := authService.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, "Token inválido", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) string {
	userID, _ := r.Context().Value(ContextUserID).(string)
	return userID
}
