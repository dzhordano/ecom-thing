package repository

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetById(ctx context.Context, orderId string) (*domain.Order, error)
	ListByUser(ctx context.Context, userId string) ([]*domain.Order, error)

	// Search(ctx context.Context, ...) ([]*domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, orderId string) error

	GetCoupon(ctx context.Context, code string) (*domain.Coupon, error)
}
