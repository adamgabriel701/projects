package main

import (
	"fmt"
	"ecommerce-api/internal/config"
	"ecommerce-api/internal/domain"
	"ecommerce-api/internal/repository"
	"ecommerce-api/internal/service"
)

func main() {
	cfg := config.Load()

	productRepo, _ := repository.NewProductRepository(cfg.DataDir)
	productService := service.NewProductService(productRepo)

	// Produtos de exemplo
	products := []domain.CreateProductRequest{
		{
			Name:        "Notebook Gamer",
			Description: "Notebook com RTX 4060, 16GB RAM, SSD 512GB",
			Price:       5999.99,
			Stock:       10,
			Category:    "Eletrônicos",
			ImageURL:    "https://example.com/notebook.jpg",
		},
		{
			Name:        "Mouse Sem Fio",
			Description: "Mouse ergonômico com DPI ajustável",
			Price:       149.99,
			Stock:       50,
			Category:    "Periféricos",
			ImageURL:    "https://example.com/mouse.jpg",
		},
		{
			Name:        "Teclado Mecânico",
			Description: "Switch Red, RGB, Layout ABNT2",
			Price:       399.99,
			Stock:       25,
			Category:    "Periféricos",
			ImageURL:    "https://example.com/teclado.jpg",
		},
		{
			Name:        "Monitor 27''",
			Description: "144Hz, 1ms, IPS, QHD",
			Price:       1899.99,
			Stock:       15,
			Category:    "Monitores",
			ImageURL:    "https://example.com/monitor.jpg",
		},
		{
			Name:        "Headset Gamer",
			Description: "Som surround 7.1, microfone removível",
			Price:       299.99,
			Stock:       30,
			Category:    "Áudio",
			ImageURL:    "https://example.com/headset.jpg",
		},
	}

	for _, p := range products {
		if _, err := productService.Create(p); err != nil {
			fmt.Printf("Erro ao criar %s: %v\n", p.Name, err)
		} else {
			fmt.Printf("✅ Criado: %s\n", p.Name)
		}
	}

	fmt.Println("\n🌱 Seed concluído!")
}
