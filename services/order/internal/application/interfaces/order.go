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

	Update(ctx context.Context /*TODO DTO*/)
	Delete(ctx context.Context, orderId uuid.UUID) error

	Search(ctx context.Context, filters map[string]any) ([]*domain.Order, error) // TODO Своя структура вместо any
	CompleteOrder(ctx context.Context, orderId uuid.UUID) error
	CancelOrder(ctx context.Context, orderId uuid.UUID) error
}
