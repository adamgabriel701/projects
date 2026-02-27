package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"ecommerce-api/internal/domain"
	"ecommerce-api/internal/middleware"
	"ecommerce-api/internal/service"
)

type CartHandler struct {
	cartService *service.CartService
}

func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

func (h *CartHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	cart := h.cartService.GetCart(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)

	var req domain.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	cart, err := h.cartService.AddItem(userID, req)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	productID := strings.TrimPrefix(r.URL.Path, "/cart/items/")

	var req domain.UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	cart, err := h.cartService.UpdateItem(userID, productID, req)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	productID := strings.TrimPrefix(r.URL.Path, "/cart/items/")

	cart, err := h.cartService.RemoveItem(userID, productID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cart)
}

func (h *CartHandler) Clear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	if err := h.cartService.ClearCart(userID); err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
