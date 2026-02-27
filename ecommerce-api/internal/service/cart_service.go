package service

import (
	"errors"
	"ecommerce-api/internal/domain"
	"ecommerce-api/internal/repository"
)

type CartService struct {
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewCartService(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *CartService) GetCart(userID string) *domain.Cart {
	return s.cartRepo.GetOrCreate(userID)
}

func (s *CartService) AddItem(userID string, req domain.AddToCartRequest) (*domain.Cart, error) {
	// Verificar produto
	product, err := s.productRepo.FindByID(req.ProductID)
	if err != nil {
		return nil, err
	}

	if product.Stock < req.Quantity {
		return nil, errors.New("estoque insuficiente")
	}

	cart := s.cartRepo.GetOrCreate(userID)

	// Verificar se item já existe no carrinho
	found := false
	for i, item := range cart.Items {
		if item.ProductID == req.ProductID {
			// Atualizar quantidade
			newQuantity := item.Quantity + req.Quantity
			if product.Stock < newQuantity {
				return nil, errors.New("estoque insuficiente para quantidade total")
			}
			cart.Items[i].Quantity = newQuantity
			cart.Items[i].Subtotal = float64(newQuantity) * item.UnitPrice
			found = true
			break
		}
	}

	if !found {
		// Adicionar novo item
		cart.Items = append(cart.Items, domain.CartItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Quantity:    req.Quantity,
			UnitPrice:   product.Price,
			Subtotal:    float64(req.Quantity) * product.Price,
		})
	}

	// Recalcular total
	s.recalculateTotal(cart)

	if err := s.cartRepo.Update(cart); err != nil {
		return nil, err
	}

	return cart, nil
}

func (s *CartService) UpdateItem(userID, productID string, req domain.UpdateCartItemRequest) (*domain.Cart, error) {
	cart := s.cartRepo.GetOrCreate(userID)

	if req.Quantity == 0 {
		// Remover item
		return s.RemoveItem(userID, productID)
	}

	// Verificar estoque
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, err
	}

	if product.Stock < req.Quantity {
		return nil, errors.New("estoque insuficiente")
	}

	// Atualizar item
	found := false
	for i, item := range cart.Items {
		if item.ProductID == productID {
			cart.Items[i].Quantity = req.Quantity
			cart.Items[i].Subtotal = float64(req.Quantity) * item.UnitPrice
			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("item não encontrado no carrinho")
	}

	s.recalculateTotal(cart)

	if err := s.cartRepo.Update(cart); err != nil {
		return nil, err
	}

	return cart, nil
}

func (s *CartService) RemoveItem(userID, productID string) (*domain.Cart, error) {
	cart := s.cartRepo.GetOrCreate(userID)

	var newItems []domain.CartItem
	for _, item := range cart.Items {
		if item.ProductID != productID {
			newItems = append(newItems, item)
		}
	}

	cart.Items = newItems
	s.recalculateTotal(cart)

	if err := s.cartRepo.Update(cart); err != nil {
		return nil, err
	}

	return cart, nil
}

func (s *CartService) ClearCart(userID string) error {
	return s.cartRepo.Clear(userID)
}

func (s *CartService) recalculateTotal(cart *domain.Cart) {
	total := 0.0
	for _, item := range cart.Items {
		total += item.Subtotal
	}
	cart.Total = total
}
