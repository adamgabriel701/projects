package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"ecommerce-api/internal/domain"
	"ecommerce-api/internal/middleware"
	"ecommerce-api/internal/service"
)

type OrderHandler struct {
	orderService   *service.OrderService
	paymentService *service.PaymentService
}

func NewOrderHandler(orderService *service.OrderService, paymentService *service.PaymentService) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		paymentService: paymentService,
	}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)

	var req domain.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	order, err := h.orderService.CreateOrder(userID, req)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	// Processar pagamento (simulação)
	var cardToken string
	if req.PaymentMethod == domain.PaymentMethodCreditCard {
		// Em produção, viria do frontend tokenizado
		cardToken = "tok_visa" // simulação
	}

	paymentResult, _ := h.paymentService.ProcessPayment(order, cardToken)
	
	if paymentResult.Success {
		// Atualizar status para pago
		order.Status = domain.OrderStatusPaid
		h.orderService.UpdateStatus(order.ID, domain.UpdateOrderStatusRequest{Status: domain.OrderStatusPaid})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order":   order,
		"payment": paymentResult,
	})
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	orders := h.orderService.GetUserOrders(userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	orderID := strings.TrimPrefix(r.URL.Path, "/orders/")

	order, err := h.orderService.GetOrder(userID, orderID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Método não permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	userID := middleware.GetUserID(r)
	orderID := strings.TrimPrefix(r.URL.Path, "/orders/")
	orderID = strings.TrimSuffix(orderID, "/cancel")

	order, err := h.orderService.CancelOrder(userID, orderID)
	if err != nil {
		http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
