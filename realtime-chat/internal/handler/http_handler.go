package handler

import (
	"encoding/json"
	"net/http"
	"realtime-chat/internal/domain"
	"realtime-chat/internal/repository"
	"realtime-chat/internal/service"
)

type HTTPHandler struct {
	hub     *service.Hub
	msgRepo *repository.MessageRepository
}

func NewHTTPHandler(hub *service.Hub, msgRepo *repository.MessageRepository) *HTTPHandler {
	return &HTTPHandler{
		hub:     hub,
		msgRepo: msgRepo,
	}
}

func (h *HTTPHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	rooms := h.hub.GetRooms()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

func (h *HTTPHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		roomID = "general"
	}

	limit := 50 // últimas 50 mensagens
	messages, err := h.msgRepo.GetRecent(roomID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.MessageResponse{
		Messages: messages,
		RoomID:   roomID,
	})
}

func (h *HTTPHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if req.ID == "" || req.Name == "" {
		http.Error(w, "ID e Nome são obrigatórios", http.StatusBadRequest)
		return
	}

	room := h.hub.CreateRoom(req.ID, req.Name, req.Description)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(domain.RoomInfo{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		UserCount:   0,
	})
}

func (h *HTTPHandler) ServeHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(HTMLTemplate))
}
