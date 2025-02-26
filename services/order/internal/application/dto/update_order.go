package dto

import (
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
)

type UpdateOrderRequest struct {
	Currency        string
	TotalPrice      float64
	PaymentMethod   string
	DeliveryMethod  string
	DeliveryAddress string
	DeliveryDate    time.Time
	Items           []domain.Item
}
