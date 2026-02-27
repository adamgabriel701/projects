package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
	"ecommerce-api/internal/domain"
)

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	return &PaymentService{}
}

type PaymentResult struct {
	Success   bool
	PaymentID string
	Message   string
}

func (s *PaymentService) ProcessPayment(order *domain.Order, cardToken string) (*PaymentResult, error) {
	// Simulação de gateway de pagamento
	// Em produção, integraria com Stripe, Pagar.me, etc.

	if order.PaymentMethod == domain.PaymentMethodPix {
		return s.processPix(order)
	}

	if order.PaymentMethod == domain.PaymentMethodBoleto {
		return s.processBoleto(order)
	}

	if order.PaymentMethod == domain.PaymentMethodCreditCard {
		return s.processCreditCard(order, cardToken)
	}

	return nil, errors.New("método de pagamento não suportado")
}

func (s *PaymentService) processPix(order *domain.Order) (*PaymentResult, error) {
	// Simular processamento PIX (aprovação instantânea em 90% dos casos)
	time.Sleep(500 * time.Millisecond)
	
	success := rand.Float32() < 0.9 // 90% de aprovação
	
	if success {
		return &PaymentResult{
			Success:   true,
			PaymentID: fmt.Sprintf("PIX-%d", rand.Int()),
			Message:   "Pagamento aprovado via PIX",
		}, nil
	}

	return &PaymentResult{
		Success: false,
		Message: "Pagamento recusado pela instituição financeira",
	}, nil
}

func (s *PaymentService) processBoleto(order *domain.Order) (*PaymentResult, error) {
	// Boleto sempre gera código, pagamento é posterior
	return &PaymentResult{
		Success:   true,
		PaymentID: fmt.Sprintf("BOL-%d", rand.Int()),
		Message:   "Boleto gerado com sucesso",
	}, nil
}

func (s *PaymentService) processCreditCard(order *domain.Order, cardToken string) (*PaymentResult, error) {
	// Simular processamento de cartão
	time.Sleep(1 * time.Second)
	
	// Simular validação do token
	if cardToken == "" {
		return &PaymentResult{
			Success: false,
			Message: "Token do cartão inválido",
		}, nil
	}

	success := rand.Float32() < 0.85 // 85% de aprovação
	
	if success {
		return &PaymentResult{
			Success:   true,
			PaymentID: fmt.Sprintf("CC-%d", rand.Int()),
			Message:   "Pagamento aprovado",
		}, nil
	}

	return &PaymentResult{
		Success: false,
		Message: "Cartão recusado. Verifique os dados e tente novamente.",
	}, nil
}
