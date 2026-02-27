package main

import (
	"fmt"
	"log"
	"net/http"
	"url-shortener/internal/handlers"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"
)

func main() {
	// Configurações
	port := "8080"
	domain := fmt.Sprintf("http://localhost:%s", port)
	
	// Inicializar dependências
	store := repository.NewMemoryStore()
	urlService := service.NewURLService(store, domain)
	urlHandler := handlers.NewURLHandler(urlService)
	
	// Configurar rotas
	mux := http.NewServeMux()
	
	// Rota para criar URL curta
	mux.HandleFunc("/shorten", urlHandler.CreateURL)
	
	// Rota para estatísticas
	mux.HandleFunc("/stats/", urlHandler.GetStats)
	
	// Rota para redirecionamento (deve ser a última)
	mux.HandleFunc("/", urlHandler.RedirectURL)
	
	// Middleware de logging
	handler := loggingMiddleware(mux)
	
	fmt.Printf("🚀 Servidor iniciado em %s\n", domain)
	fmt.Println("📚 Endpoints disponíveis:")
	fmt.Println("   POST /shorten     - Criar URL curta")
	fmt.Println("   GET  /stats/{code}- Estatísticas da URL")
	fmt.Println("   GET  /{code}      - Redirecionar para URL original")
	
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s - %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
