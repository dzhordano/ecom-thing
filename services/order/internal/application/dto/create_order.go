package dto

import (
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
)

type CreateOrderRequest struct {
	Currency        string
	TotalPrice      float64
	Coupon          string
	PaymentMethod   string
	DeliveryMethod  string
	DeliveryAddress string
	DeliveryDate    time.Time
	Items           []domain.Item
}
