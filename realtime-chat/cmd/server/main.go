package main

import (
	"fmt"
	"log"
	"net/http"
	"realtime-chat/internal/handler"
	"realtime-chat/internal/repository"
	"realtime-chat/internal/service"
)

func main() {
	// Inicializar repositório
	msgRepo, err := repository.NewMessageRepository("./data")
	if err != nil {
		log.Fatal("Erro ao inicializar repositório:", err)
	}

	// Inicializar hub
	hub := service.NewHub(msgRepo)
	go hub.Run()

	// Inicializar handlers
	wsHandler := handler.NewWebSocketHandler(hub)
	httpHandler := handler.NewHTTPHandler(hub, msgRepo)

	// Configurar rotas
	mux := http.NewServeMux()
	
	// WebSocket
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)
	
	// API REST
	mux.HandleFunc("/api/rooms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			httpHandler.GetRooms(w, r)
		} else if r.Method == http.MethodPost {
			httpHandler.CreateRoom(w, r)
		}
	})
	mux.HandleFunc("/api/messages", httpHandler.GetMessages)
	
	// Interface web
	mux.HandleFunc("/", httpHandler.ServeHTML)

	// Middleware CORS e logging
	handler := corsMiddleware(loggingMiddleware(mux))

	port := "8080"
	fmt.Printf("🚀 Chat Server iniciado em http://localhost:%s\n", port)
	fmt.Println("\n📚 Endpoints:")
	fmt.Println("   WebSocket: ws://localhost:%s/ws?room=general&username=Joao")
	fmt.Println("   GET  /api/rooms     - Listar salas")
	fmt.Println("   POST /api/rooms     - Criar sala")
	fmt.Println("   GET  /api/messages  - Histórico de mensagens")

	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s - %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
