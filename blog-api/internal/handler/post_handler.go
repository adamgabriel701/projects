package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"blog-api/internal/domain"
	"blog-api/internal/middleware"
	"blog-api/internal/service"
	"blog-api/internal/utils"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Parse filtros
	filter := domain.PostFilter{
		Status: domain.StatusPublished, // default: apenas publicados
	}

	if category := r.URL.Query().Get("category"); category != "" {
		filter.CategoryID = category
	}
	if tag := r.URL.Query().Get("tag"); tag != "" {
		filter.Tag = tag
	}
	if author := r.URL.Query().Get("author"); author != "" {
		filter.AuthorID = author
	}
	if featured := r.URL.Query().Get("featured"); featured == "true" {
		filter.Featured = true
	}
	if status := r.URL.Query().Get("status"); status != "" {
		// Apenas admins podem ver posts não publicados
		userRole := middleware.GetUserRole(r)
		if userRole == domain.RoleAdmin || userRole == domain.RoleEditor {
			filter.Status = domain.PostStatus(status)
		}
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}
	filter.Page = page
	filter.Limit = limit

	posts, total, err := h.postService.List(filter)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := (total + limit - 1) / limit

	utils.SendSuccess(w, domain.PostListResponse{
		Posts:      posts,
		Total:      total,
		Page:       page,
		TotalPages: totalPages,
	}, http.StatusOK)
}

func (h *PostHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/posts/")
	incrementViews := r.URL.Query().Get("view") != "false"

	post, err := h.postService.GetBySlug(slug, incrementViews)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusNotFound)
		return
	}

	utils.SendSuccess(w, post, http.StatusOK)
}

func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == "" {
		utils.SendError(w, "não autorizado", http.StatusUnauthorized)
		return
	}

	var req domain.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	post, err := h.postService.Create(userID, req)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendSuccess(w, post, http.StatusCreated)
}

func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	userRole := middleware.GetUserRole(r)

	postID := strings.TrimPrefix(r.URL.Path, "/posts/")
	postID = strings.TrimSuffix(postID, "/update")

	var req domain.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	post, err := h.postService.Update(userID, userRole, postID, req)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendSuccess(w, post, http.StatusOK)
}

func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	userRole := middleware.GetUserRole(r)

	postID := strings.TrimPrefix(r.URL.Path, "/posts/")

	if err := h.postService.Delete(userID, userRole, postID); err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PostHandler) ToggleLike(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if userID == "" {
		utils.SendError(w, "não autorizado", http.StatusUnauthorized)
		return
	}

	postID := strings.TrimPrefix(r.URL.Path, "/posts/")
	postID = strings.TrimSuffix(postID, "/like")

	liked, err := h.postService.ToggleLike(postID, userID)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendSuccess(w, map[string]bool{"liked": liked}, http.StatusOK)
}

func (h *PostHandler) GetFeatured(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	posts, err := h.postService.GetFeatured()
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendSuccess(w, posts, http.StatusOK)
}
