package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"blog-api/internal/domain"
	"blog-api/internal/middleware"
	"blog-api/internal/service"
	"blog-api/internal/utils"
)

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r) // pode ser vazio para anônimos

	var req domain.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	comment, err := h.commentService.Create(userID, req)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendSuccess(w, comment, http.StatusCreated)
}

func (h *CommentHandler) GetByPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	postID := strings.TrimPrefix(r.URL.Path, "/posts/")
	postID = strings.TrimSuffix(postID, "/comments")

	userRole := middleware.GetUserRole(r)
	includePending := userRole == domain.RoleEditor || userRole == domain.RoleAdmin

	comments, err := h.commentService.GetByPost(postID, includePending)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusNotFound)
		return
	}

	utils.SendSuccess(w, comments, http.StatusOK)
}

func (h *CommentHandler) Moderate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	userRole := middleware.GetUserRole(r)

	commentID := strings.TrimPrefix(r.URL.Path, "/comments/")
	commentID = strings.TrimSuffix(commentID, "/moderate")

	var req domain.UpdateCommentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.commentService.Moderate(userID, userRole, commentID, req.Status); err != nil {
		utils.SendError(w, err.Error(), http.StatusForbidden)
		return
	}

	utils.SendSuccess(w, map[string]string{"status": "updated"}, http.StatusOK)
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	userRole := middleware.GetUserRole(r)

	commentID := strings.TrimPrefix(r.URL.Path, "/comments/")

	if err := h.commentService.Delete(userID, userRole, commentID); err != nil {
		utils.SendError(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
