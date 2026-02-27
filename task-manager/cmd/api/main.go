package main

import (
	"fmt"
	"log"
	"net/http"
	"task-manager/internal/config"
	"task-manager/internal/handler"
	"task-manager/internal/middleware"
	"task-manager/internal/repository"
	"task-manager/internal/service"
)

func main() {
	cfg := config.Load()

	// Inicializar repositórios
	userRepo, err := repository.NewUserRepository(cfg.DataDir)
	if err != nil {
		log.Fatal("Erro ao inicializar repositório de usuários:", err)
	}

	taskRepo, err := repository.NewTaskRepository(cfg.DataDir)
	if err != nil {
		log.Fatal("Erro ao inicializar repositório de tarefas:", err)
	}

	// Inicializar serviços
	authService := service.NewAuthService(userRepo, cfg)
	taskService := service.NewTaskService(taskRepo)

	// Inicializar handlers
	authHandler := handler.NewAuthHandler(authService)
	taskHandler := handler.NewTaskHandler(taskService)

	// Configurar rotas
	mux := http.NewServeMux()

	// Rotas públicas
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)

	// Rotas protegidas (com middleware)
	authMiddleware := middleware.AuthMiddleware(authService)
	
	// Tasks routes
	mux.Handle("/tasks", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			taskHandler.List(w, r)
		case http.MethodPost:
			taskHandler.Create(w, r)
		default:
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/tasks/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			taskHandler.Update(w, r)
		case http.MethodDelete:
			taskHandler.Delete(w, r)
		default:
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		}
	})))

	// Middleware de logging
	handler := loggingMiddleware(mux)

	fmt.Printf("🚀 Task Manager API iniciada em http://localhost:%s\n", cfg.Port)
	fmt.Println("\n📚 Endpoints:")
	fmt.Println("  Públicos:")
	fmt.Println("    POST /auth/register  - Registrar usuário")
	fmt.Println("    POST /auth/login     - Login")
	fmt.Println("\n  Protegidos (requer Bearer token):")
	fmt.Println("    GET    /tasks        - Listar tarefas (?status=pending|in_progress|completed)")
	fmt.Println("    POST   /tasks        - Criar tarefa")
	fmt.Println("    PUT    /tasks/{id}   - Atualizar tarefa")
	fmt.Println("    DELETE /tasks/{id}   - Deletar tarefa")

	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s - %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
