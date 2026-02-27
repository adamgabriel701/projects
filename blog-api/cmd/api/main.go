package main

import (
	"blog-api/internal/config"
	"blog-api/internal/handler"
	"blog-api/internal/middleware"
	"blog-api/internal/repository"
	"blog-api/internal/service"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	cfg := config.Load()

	// Inicializar repositórios
	userRepo, err := repository.NewUserRepository(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}

	postRepo, err := repository.NewPostRepository(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}

	commentRepo, err := repository.NewCommentRepository(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}

	categoryRepo, err := repository.NewCategoryRepository(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}

	tagRepo, err := repository.NewTagRepository(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}

	// Inicializar serviços
	authService := service.NewAuthService(userRepo, cfg)
	postService := service.NewPostService(postRepo, categoryRepo, tagRepo)
	commentService := service.NewCommentService(commentRepo, postRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	searchService := service.NewSearchService(postRepo)

	// Inicializar handlers
	authHandler := handler.NewAuthHandler(authService)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	searchHandler := handler.NewSearchHandler(searchService)

	// Middleware
	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)
	optionalAuthMiddleware := middleware.OptionalAuthMiddleware(cfg.JWTSecret)

	// Configurar rotas
	mux := http.NewServeMux()

	// Rotas públicas
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)

	// Posts públicos
	mux.Handle("/posts", optionalAuthMiddleware(http.HandlerFunc(postHandler.List)))
	mux.Handle("/posts/featured", http.HandlerFunc(postHandler.GetFeatured))
	mux.Handle("/posts/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Roteamento dinâmico para /posts/{slug} e /posts/{id}/comments
		path := r.URL.Path

		if strings.HasSuffix(path, "/comments") {
			commentHandler.GetByPost(w, r)
		} else if strings.HasSuffix(path, "/like") {
			authMiddleware(http.HandlerFunc(postHandler.ToggleLike)).ServeHTTP(w, r)
		} else if strings.HasSuffix(path, "/update") {
			authMiddleware(http.HandlerFunc(postHandler.Update)).ServeHTTP(w, r)
		} else {
			postHandler.Get(w, r)
		}
	}))

	// Categorias
	mux.Handle("/categories", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authMiddleware(http.HandlerFunc(categoryHandler.Create)).ServeHTTP(w, r)
		} else {
			categoryHandler.List(w, r)
		}
	}))
	mux.Handle("/categories/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/update") {
			authMiddleware(http.HandlerFunc(categoryHandler.Update)).ServeHTTP(w, r)
		} else {
			switch r.Method {
			case http.MethodGet:
				categoryHandler.Get(w, r)
			case http.MethodDelete:
				authMiddleware(http.HandlerFunc(categoryHandler.Delete)).ServeHTTP(w, r)
			}
		}
	}))

	// Busca
	mux.HandleFunc("/search", searchHandler.Search)
	mux.HandleFunc("/search/related", searchHandler.Related)

	// Rotas protegidas (requerem autenticação)
	mux.Handle("/auth/profile", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			authHandler.GetProfile(w, r)
		case http.MethodPut:
			authHandler.UpdateProfile(w, r)
		}
	})))

	mux.Handle("/posts/create", authMiddleware(http.HandlerFunc(postHandler.Create)))

	// Comentários
	mux.Handle("/comments", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			optionalAuthMiddleware(http.HandlerFunc(commentHandler.Create)).ServeHTTP(w, r)
		}
	}))
	mux.Handle("/comments/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/moderate") {
			authMiddleware(http.HandlerFunc(commentHandler.Moderate)).ServeHTTP(w, r)
		} else if r.Method == http.MethodDelete {
			authMiddleware(http.HandlerFunc(commentHandler.Delete)).ServeHTTP(w, r)
		}
	}))

	// Middleware global
	handler := corsMiddleware(middleware.LoggingMiddleware(mux))

	port := cfg.Port
	fmt.Printf("🚀 Blog API iniciado em http://localhost:%s\n\n", port)

	fmt.Println("📚 Endpoints:")
	fmt.Println("\n🔐 Autenticação:")
	fmt.Println("  POST /auth/register       - Registrar")
	fmt.Println("  POST /auth/login          - Login")
	fmt.Println("  GET  /auth/profile        - Perfil (auth)")
	fmt.Println("  PUT  /auth/profile        - Atualizar perfil (auth)")

	fmt.Println("\n📝 Posts:")
	fmt.Println("  GET  /posts               - Listar posts")
	fmt.Println("  GET  /posts/featured      - Posts em destaque")
	fmt.Println("  GET  /posts/{slug}        - Ver post")
	fmt.Println("  POST /posts/create        - Criar post (auth)")
	fmt.Println("  PUT  /posts/{id}/update   - Atualizar post (auth)")
	fmt.Println("  DELETE /posts/{id}        - Deletar post (auth)")
	fmt.Println("  POST /posts/{id}/like     - Like/unlike (auth)")

	fmt.Println("\n💬 Comentários:")
	fmt.Println("  GET  /posts/{id}/comments - Listar comentários")
	fmt.Println("  POST /comments            - Criar comentário")
	fmt.Println("  PUT  /comments/{id}/moderate - Moderar (editor/admin)")
	fmt.Println("  DELETE /comments/{id}     - Deletar comentário")

	fmt.Println("\n🏷️ Categorias:")
	fmt.Println("  GET  /categories          - Listar categorias")
	fmt.Println("  GET  /categories/{slug}   - Ver categoria")
	fmt.Println("  POST /categories          - Criar categoria (admin)")
	fmt.Println("  PUT  /categories/{id}/update - Atualizar (admin)")
	fmt.Println("  DELETE /categories/{id}   - Deletar (admin)")

	fmt.Println("\n🔍 Busca:")
	fmt.Println("  GET  /search?q=termo      - Buscar posts")
	fmt.Println("  GET  /search/related?post_id=xxx - Posts relacionados")

	log.Fatal(http.ListenAndServe(":"+port, handler))
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
