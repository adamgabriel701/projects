package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"realtime-chat/internal/domain"
	"realtime-chat/internal/service"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Permitir todas as origens em desenvolvimento
	},
}

type WebSocketHandler struct {
	hub *service.Hub
}

func NewWebSocketHandler(hub *service.Hub) *WebSocketHandler {
	return &WebSocketHandler{hub: hub}
}

func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Pegar parâmetros da query
	roomID := r.URL.Query().Get("room")
	username := r.URL.Query().Get("username")

	if roomID == "" {
		roomID = "general"
	}
	if username == "" {
		username = "Anônimo"
	}

	// Verificar se sala existe
	if h.hub.GetRoom(roomID) == nil {
		http.Error(w, "Sala não encontrada", http.StatusNotFound)
		return
	}

	// Upgrade para WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro no upgrade WebSocket: %v", err)
		return
	}

	client := &domain.Client{
		ID:       generateID(),
		Username: username,
		RoomID:   roomID,
		Send:     make(chan domain.Message, 256),
	}

	// Registrar no hub
	h.hub.GetRegisterChan() <- client

	// Iniciar goroutines
	go h.writePump(client, conn)
	go h.readPump(client, conn)
}

func (h *WebSocketHandler) readPump(client *domain.Client, conn *websocket.Conn) {
	defer func() {
		h.hub.GetUnregisterChan() <- client
		conn.Close()
	}()

	conn.SetReadLimit(512 * 1024) // 512KB max message size

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Erro de leitura WebSocket: %v", err)
			}
			break
		}

		var msgData struct {
			Content string `json:"content"`
		}
		
		if err := json.Unmarshal(message, &msgData); err != nil {
			continue
		}

		msg := domain.Message{
			ID:        generateID(),
			RoomID:    client.RoomID,
			Username:  client.Username,
			Content:   msgData.Content,
			Type:      "text",
			CreatedAt: time.Now(),
		}

		// Enviar diretamente para o broadcast do hub
		h.hub.GetBroadcastChan() <- msg
	}
}

func (h *WebSocketHandler) writePump(client *domain.Client, conn *websocket.Conn) {
	defer conn.Close()

	for msg := range client.Send {
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			break
		}
	}
}

func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}