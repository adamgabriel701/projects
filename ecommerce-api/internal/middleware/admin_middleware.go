package middleware

import (
	"net/http"
	"ecommerce-api/internal/domain"
)

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := GetUserRole(r)
		if role != domain.RoleAdmin {
			http.Error(w, `{"error": "Acesso restrito a administradores"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
