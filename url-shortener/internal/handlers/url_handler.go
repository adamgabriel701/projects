package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"url-shortener/internal/models"
	"url-shortener/internal/service"
)

type URLHandler struct {
	service *service.URLService
}

func NewURLHandler(service *service.URLService) *URLHandler {
	return &URLHandler{service: service}
}

// Criar URL curta
func (h *URLHandler) CreateURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	
	var req models.CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	
	if req.URL == "" {
		http.Error(w, "URL é obrigatória", http.StatusBadRequest)
		return
	}
	
	// Adicionar http:// se não tiver protocolo
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		req.URL = "https://" + req.URL
	}
	
	resp, err := h.service.CreateShortURL(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// Redirecionar para URL original
func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	
	// Extrair código curto da URL /{code}
	shortCode := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == "" {
		http.Error(w, "Código não fornecido", http.StatusBadRequest)
		return
	}
	
	originalURL, err := h.service.GetOriginalURL(shortCode)
	if err != nil {
		http.Error(w, "URL não encontrada", http.StatusNotFound)
		return
	}
	
	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

// Estatísticas da URL
func (h *URLHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	
	// Extrair código: /stats/{code}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/stats/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Código não fornecido", http.StatusBadRequest)
		return
	}
	
	url, err := h.service.GetURLStats(parts[0])
	if err != nil {
		http.Error(w, "URL não encontrada", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(url)
}
