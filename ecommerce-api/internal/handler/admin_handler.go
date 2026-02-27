package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"ecommerce-api/internal/domain"
	"ecommerce-api/internal/service"
)

type AdminHandler struct {
	orderService   *service.OrderService
	productService *service.ProductService
}

func NewAdminHandler(orderService *service.OrderService, productService *service.ProductService) *AdminHandler {
	return &AdminHandler{
		orderService:   orderService,
		productService: productService,
	}
}

func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	status := domain.OrderStatus(r.URL.Query().Get("status"))
	
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 20
	}

	orders, total := h.orderService.GetAllOrders(status, page, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"orders": orders,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

func (h *AdminHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	orderID := r.URL.Path[len("/admin/orders/"):]
	orderID = orderID[:len(orderID)-len("/status")]

	var req domain.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	order, err := h.orderService.UpdateStatus(orderID, req)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	// Estatísticas simples
	orders, _ := h.orderService.GetAllOrders("", 1, 1000)
	
	totalSales := 0.0
	statusCount := make(map[domain.OrderStatus]int)
	
	for _, order := range orders {
		if order.Status == domain.OrderStatusDelivered || order.Status == domain.OrderStatusPaid {
			totalSales += order.Total
		}
		statusCount[order.Status]++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_orders": len(orders),
		"total_sales":  totalSales,
		"by_status":    statusCount,
	})
}
