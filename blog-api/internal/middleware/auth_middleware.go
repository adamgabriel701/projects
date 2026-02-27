package middleware

import (
	"context"
	"net/http"
	"strings"
	"blog-api/internal/domain"
	"blog-api/internal/utils"
)

type contextKey string

const (
	ContextUserID   contextKey = "user_id"
	ContextUserRole contextKey = "user_role"
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "token não fornecido"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error": "formato de token inválido"}`, http.StatusUnauthorized)
				return
			}

			claims, err := utils.ValidateToken(parts[1], secret)
			if err != nil {
				http.Error(w, `{"error": "token inválido"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextUserRole, domain.UserRole(claims.Role))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := utils.ValidateToken(parts[1], secret)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextUserRole, domain.UserRole(claims.Role))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(r *http.Request) string {
	id, _ := r.Context().Value(ContextUserID).(string)
	return id
}

func GetUserRole(r *http.Request) domain.UserRole {
	role, _ := r.Context().Value(ContextUserRole).(domain.UserRole)
	return role
}
