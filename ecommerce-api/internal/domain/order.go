package domain

import "time"

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodPix        PaymentMethod = "pix"
	PaymentMethodBoleto     PaymentMethod = "boleto"
)

type Order struct {
	ID            string        `json:"id"`
	UserID        string        `json:"user_id"`
	Items         []OrderItem   `json:"items"`
	Total         float64       `json:"total"`
	Status        OrderStatus   `json:"status"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	PaymentID     string        `json:"payment_id,omitempty"`
	Shipping      ShippingInfo  `json:"shipping"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Subtotal    float64 `json:"subtotal"`
}

type ShippingInfo struct {
	Address   Address `json:"address"`
	Cost      float64 `json:"cost"`
	Tracking  string  `json:"tracking,omitempty"`
	Estimated time.Time `json:"estimated_delivery"`
}

type CreateOrderRequest struct {
	PaymentMethod PaymentMethod `json:"payment_method" validate:"required"`
	ShippingAddress Address     `json:"shipping_address" validate:"required"`
}

type OrderResponse struct {
	ID            string        `json:"id"`
	Items         []OrderItem   `json:"items"`
	Total         float64       `json:"total"`
	Status        OrderStatus   `json:"status"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Shipping      ShippingInfo  `json:"shipping"`
	CreatedAt     time.Time     `json:"created_at"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" validate:"required"`
	Tracking string    `json:"tracking,omitempty"`
}
