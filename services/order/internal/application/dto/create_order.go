package dto

import (
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	order_v1 "github.com/dzhordano/ecom-thing/services/order/pkg/api/order/v1"
	"github.com/google/uuid"
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

func RPCItemsToDomain(items []*order_v1.Item) ([]domain.Item, error) {
	var result []domain.Item
	for _, item := range items {
		id, err := uuid.Parse(item.ItemId)
		if err != nil {
			return nil, domain.ErrInvalidUUID // FIXME too obscure
		}

		result = append(result, domain.Item{
			ProductID: id,
			Quantity:  item.GetQuantity(),
		})
	}
	return result, nil
}
