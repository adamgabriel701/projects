package handler

import (
	"encoding/json"
	"net/http"
	"blog-api/internal/domain"
	"blog-api/internal/middleware"
	"blog-api/internal/service"
	"blog-api/internal/utils"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	user, token, err := h.authService.Register(req)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendSuccess(w, map[string]interface{}{
		"token": token,
		"user": domain.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			Name:      user.Name,
			Bio:       user.Bio,
			Avatar:    user.Avatar,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	}, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	user, token, err := h.authService.Login(req)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	utils.SendSuccess(w, map[string]interface{}{
		"token": token,
		"user": domain.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			Name:      user.Name,
			Bio:       user.Bio,
			Avatar:    user.Avatar,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	}, http.StatusOK)
}

func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	user, err := h.authService.GetProfile(userID)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusNotFound)
		return
	}

	utils.SendSuccess(w, domain.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Name:      user.Name,
		Bio:       user.Bio,
		Avatar:    user.Avatar,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, http.StatusOK)
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)

	var req domain.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	user, err := h.authService.UpdateProfile(userID, req)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendSuccess(w, domain.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		Name:      user.Name,
		Bio:       user.Bio,
		Avatar:    user.Avatar,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, http.StatusOK)
}
