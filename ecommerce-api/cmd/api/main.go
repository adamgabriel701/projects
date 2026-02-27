package main

import (
	"fmt"
	"log"
	"net/http"
	"ecommerce-api/internal/config"
	"ecommerce-api/internal/handler"
	"ecommerce-api/internal/middleware"
	"ecommerce-api/internal/repository"
	"ecommerce-api/internal/service"
)

func main() {
	cfg := config.Load()

	// Inicializar repositórios
	userRepo, err := repository.NewUserRepository(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}

	productRepo, err := repository.NewProductRepository(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}

	cartRepo, err := repository.NewCartRepository(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}

	orderRepo, err := repository.NewOrderRepository(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}

	// Inicializar serviços
	authService := service.NewAuthService(userRepo, cfg)
	productService := service.NewProductService(productRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	orderService := service.NewOrderService(orderRepo, cartRepo, productRepo, cartService)
	paymentService := service.NewPaymentService()

	// Criar admin padrão
	if err := authService.CreateAdmin(); err != nil {
		log.Printf("Admin já existe ou erro: %v", err)
	}

	// Inicializar handlers
	authHandler := handler.NewAuthHandler(authService)
	productHandler := handler.NewProductHandler(productService)
	cartHandler := handler.NewCartHandler(cartService)
	orderHandler := handler.NewOrderHandler(orderService, paymentService)
	adminHandler := handler.NewAdminHandler(orderService, productService)

	// Middleware
	authMiddleware := middleware.AuthMiddleware(authService)

	// Configurar rotas
	mux := http.NewServeMux()

	// Rotas públicas
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/products", productHandler.List)
	mux.HandleFunc("/products/", productHandler.Get)

	// Rotas protegidas (usuários logados)
	mux.Handle("/cart", authMiddleware(http.HandlerFunc(cartHandler.Get)))
	mux.Handle("/cart/items", authMiddleware(http.HandlerFunc(cartHandler.AddItem)))
	mux.Handle("/cart/items/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			cartHandler.UpdateItem(w, r)
		case http.MethodDelete:
			cartHandler.RemoveItem(w, r)
		default:
			http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/cart/clear", authMiddleware(http.HandlerFunc(cartHandler.Clear)))

	mux.Handle("/orders", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			orderHandler.Create(w, r)
		} else if r.Method == http.MethodGet {
			orderHandler.List(w, r)
		}
	})))
	mux.Handle("/orders/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			orderHandler.Get(w, r)
		} else if r.Method == http.MethodPost {
			orderHandler.Cancel(w, r)
		}
	})))

	// Rotas administrativas
	adminRoutes := authMiddleware(middleware.AdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		
		switch {
		case path == "/admin/dashboard":
			adminHandler.Dashboard(w, r)
		case path == "/admin/orders":
			adminHandler.ListOrders(w, r)
		case len(path) > len("/admin/orders/") && path[len(path)-7:] == "/status":
			adminHandler.UpdateOrderStatus(w, r)
		case path == "/admin/products":
			productHandler.Create(w, r)
		case len(path) > len("/admin/products/"):
			if r.Method == http.MethodPut {
				productHandler.Update(w, r)
			} else if r.Method == http.MethodDelete {
				productHandler.Delete(w, r)
			}
		default:
			http.Error(w, `{"error": "Rota não encontrada"}`, http.StatusNotFound)
		}
	})))
	mux.Handle("/admin/", adminRoutes)

	// Middleware global
	handler := corsMiddleware(loggingMiddleware(mux))

	port := cfg.Port
	fmt.Printf("🚀 E-commerce API iniciado em http://localhost:%s\n\n", port)
	
	fmt.Println("📚 Endpoints Públicos:")
	fmt.Println("  POST /auth/register       - Registrar usuário")
	fmt.Println("  POST /auth/login          - Login")
	fmt.Println("  GET  /products            - Listar produtos")
	fmt.Println("  GET  /products/{id}       - Detalhes do produto")
	
	fmt.Println("\n🔒 Endpoints Protegidos (requer token):")
	fmt.Println("  GET  /cart                - Ver carrinho")
	fmt.Println("  POST /cart/items          - Adicionar ao carrinho")
	fmt.Println("  PUT  /cart/items/{id}     - Atualizar quantidade")
	fmt.Println("  DELETE /cart/items/{id}   - Remover item")
	fmt.Println("  POST /orders              - Criar pedido")
	fmt.Println("  GET  /orders              - Meus pedidos")
	fmt.Println("  GET  /orders/{id}         - Detalhes do pedido")
	fmt.Println("  POST /orders/{id}/cancel  - Cancelar pedido")
	
	fmt.Println("\n👑 Endpoints Admin:")
	fmt.Println("  GET  /admin/dashboard     - Estatísticas")
	fmt.Println("  GET  /admin/orders        - Listar todos pedidos")
	fmt.Println("  PUT  /admin/orders/{id}/status - Atualizar status")
	fmt.Println("  POST /admin/products      - Criar produto")
	fmt.Println("  PUT  /admin/products/{id} - Atualizar produto")
	fmt.Println("  DELETE /admin/products/{id} - Deletar produto")
	
	fmt.Println("\n🔑 Credenciais Admin:")
	fmt.Printf("  Email: %s\n", cfg.AdminEmail)
	fmt.Printf("  Senha: %s\n", cfg.AdminPassword)

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
