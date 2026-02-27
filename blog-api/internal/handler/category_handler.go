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

type CategoryHandler struct {
	categoryService *service.CategoryService
}

func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	categories := h.categoryService.GetAll()
	utils.SendSuccess(w, categories, http.StatusOK)
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/categories/")
	category, err := h.categoryService.GetBySlug(slug)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusNotFound)
		return
	}

	utils.SendSuccess(w, category, http.StatusOK)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Verificar se é admin
	userRole := middleware.GetUserRole(r)
	if userRole != domain.RoleAdmin {
		utils.SendError(w, "não autorizado", http.StatusForbidden)
		return
	}

	var req domain.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.Create(req)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendSuccess(w, category, http.StatusCreated)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userRole := middleware.GetUserRole(r)
	if userRole != domain.RoleAdmin {
		utils.SendError(w, "não autorizado", http.StatusForbidden)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/categories/")
	id = strings.TrimSuffix(id, "/update")

	var req domain.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	category, err := h.categoryService.Update(id, req)
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendSuccess(w, category, http.StatusOK)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.SendError(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userRole := middleware.GetUserRole(r)
	if userRole != domain.RoleAdmin {
		utils.SendError(w, "não autorizado", http.StatusForbidden)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/categories/")

	if err := h.categoryService.Delete(id); err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
