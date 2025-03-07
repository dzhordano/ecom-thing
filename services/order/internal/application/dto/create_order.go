package dto

import (
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
)

type CreateOrderRequest struct {
	Description     string
	Currency        string
	Coupon          string
	PaymentMethod   string
	DeliveryMethod  string
	DeliveryAddress string
	DeliveryDate    time.Time
	Items           []domain.Item
}
