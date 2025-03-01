package interfaces

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(ctx context.Context, info dto.CreateOrderRequest) (*domain.Order, error)

	GetById(ctx context.Context, orderId uuid.UUID) (*domain.Order, error)
	ListByUser(ctx context.Context, limit, offset uint64) ([]*domain.Order, error)

	UpdateOrder(ctx context.Context, info dto.UpdateOrderRequest) (*domain.Order, error)
	DeleteOrder(ctx context.Context, orderId uuid.UUID) error

	SearchOrders(ctx context.Context, filters map[string]any) ([]*domain.Order, error) // TODO Своя структура вместо any
	CompleteOrder(ctx context.Context, orderId uuid.UUID) error
	CancelOrder(ctx context.Context, orderId uuid.UUID) error
}

type ProductService interface {
	GetProductInfo(ctx context.Context, orderId uuid.UUID) (float64, bool, error)
}

type InventoryService interface {
	// Lock product quantity.
	ReserveQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error
	// Release reserved product quantity
	ReleaseQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error
	// Release and subtract quantity.
	SubReservedQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error
}
