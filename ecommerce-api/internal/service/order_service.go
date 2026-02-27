package service

import (
	"errors"
	"time"
	"ecommerce-api/internal/domain"
	"ecommerce-api/internal/repository"

	"github.com/google/uuid"
)

type OrderService struct {
	orderRepo   *repository.OrderRepository
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
	cartService *CartService
}

func NewOrderService(orderRepo *repository.OrderRepository, cartRepo *repository.CartRepository, productRepo *repository.ProductRepository, cartService *CartService) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
		cartService: cartService,
	}
}

func (s *OrderService) CreateOrder(userID string, req domain.CreateOrderRequest) (*domain.Order, error) {
	// Pegar carrinho
	cart := s.cartRepo.GetOrCreate(userID)
	if len(cart.Items) == 0 {
		return nil, errors.New("carrinho vazio")
	}

	// Verificar estoque de todos os itens
	for _, item := range cart.Items {
		if err := s.productRepo.UpdateStock(item.ProductID, 0); err != nil { // apenas verifica existência
			return nil, err
		}
		product, _ := s.productRepo.FindByID(item.ProductID)
		if product.Stock < item.Quantity {
			return nil, errors.New("produto " + product.Name + " sem estoque suficiente")
		}
	}

	// Criar itens do pedido
	var orderItems []domain.OrderItem
	var total float64

	for _, item := range cart.Items {
		orderItems = append(orderItems, domain.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    item.Subtotal,
		})
		total += item.Subtotal

		// Decrementar estoque
		if err := s.productRepo.UpdateStock(item.ProductID, -item.Quantity); err != nil {
			return nil, err
		}
	}

	// Calcular frete (simplificado)
	shippingCost := 15.0
	if total > 200 {
		shippingCost = 0 // frete grátis acima de 200
	}
	total += shippingCost

	order := &domain.Order{
		ID:            uuid.New().String(),
		UserID:        userID,
		Items:         orderItems,
		Total:         total,
		Status:        domain.OrderStatusPending,
		PaymentMethod: req.PaymentMethod,
		Shipping: domain.ShippingInfo{
			Address:   req.ShippingAddress,
			Cost:      shippingCost,
			Estimated: time.Now().Add(7 * 24 * time.Hour), // 7 dias úteis
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	// Limpar carrinho
	s.cartService.ClearCart(userID)

	return order, nil
}

func (s *OrderService) GetOrder(userID, orderID string) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}

	// Verificar se o pedido pertence ao usuário (ou é admin)
	if order.UserID != userID {
		return nil, errors.New("não autorizado")
	}

	return order, nil
}

func (s *OrderService) GetUserOrders(userID string) []domain.Order {
	return s.orderRepo.FindByUser(userID)
}

func (s *OrderService) GetAllOrders(status domain.OrderStatus, page, limit int) ([]domain.Order, int) {
	return s.orderRepo.FindAll(status, page, limit)
}

func (s *OrderService) UpdateStatus(orderID string, req domain.UpdateOrderStatusRequest) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}

	order.Status = req.Status
	if req.Tracking != "" {
		order.Shipping.Tracking = req.Tracking
	}
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) CancelOrder(userID, orderID string) (*domain.Order, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, errors.New("não autorizado")
	}

	if order.Status != domain.OrderStatusPending && order.Status != domain.OrderStatusPaid {
		return nil, errors.New("pedido não pode ser cancelado neste status")
	}

	// Devolver estoque
	for _, item := range order.Items {
		s.productRepo.UpdateStock(item.ProductID, item.Quantity)
	}

	order.Status = domain.OrderStatusCancelled
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(order); err != nil {
		return nil, err
	}

	return order, nil
}
