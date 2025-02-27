package dto

import (
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"github.com/google/uuid"
)

type UpdateOrderRequest struct {
	OrderID         uuid.UUID
	Status          *string
	TotalPrice      *float64
	PaymentMethod   *string
	DeliveryMethod  *string
	DeliveryAddress *string
	DeliveryDate    time.Time
	Items           []domain.Item
}
