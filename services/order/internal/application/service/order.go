package service

import (
	"context"
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain/repository"
	"github.com/google/uuid"
)

type OrderService struct {
	repo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) interfaces.OrderService {
	return &OrderService{
		repo: repo,
	}
}

// CreateOrder implements interfaces.OrderService.
func (o *OrderService) CreateOrder(ctx context.Context, info dto.CreateOrderRequest) (*domain.Order, error) {
	disc, err := o.repo.GetCoupon(ctx, info.Coupon)
	if err != nil {
		return nil, err
	}

	if disc.ValidTo.Before(time.Now()) {
		return nil, domain.ErrCouponExpired
	}

	order, err := domain.NewOrder(
		uuid.New(),
		domain.OrderPending.String(),
		info.Currency,
		info.TotalPrice,
		disc.Discount,
		info.PaymentMethod,
		info.DeliveryMethod,
		info.DeliveryAddress,
		info.DeliveryDate,
		info.Items,
	)
	if err != nil {
		return nil, err
	}

	if err := o.repo.Create(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// Delete implements interfaces.OrderService.
func (o *OrderService) Delete(ctx context.Context, orderId uuid.UUID) error {
	panic("unimplemented")
}

// GetById implements interfaces.OrderService.
func (o *OrderService) GetById(ctx context.Context, orderId uuid.UUID) (*domain.Order, error) {
	panic("unimplemented")
}

// ListByUser implements interfaces.OrderService.
func (o *OrderService) ListByUser(ctx context.Context, limit uint64, offset uint64) ([]*domain.Order, error) {
	panic("unimplemented")
}

// Search implements interfaces.OrderService.
func (o *OrderService) Search(ctx context.Context, filters map[string]any) ([]*domain.Order, error) {
	panic("unimplemented")
}

// Update implements interfaces.OrderService.
func (o *OrderService) Update(ctx context.Context) {
	panic("unimplemented")
}

// CancelOrder implements interfaces.OrderService.
func (o *OrderService) CancelOrder(ctx context.Context, orderId uuid.UUID) error {
	panic("unimplemented")
}

// CompleteOrder implements interfaces.OrderService.
func (o *OrderService) CompleteOrder(ctx context.Context, orderId uuid.UUID) error {
	panic("unimplemented")
}
