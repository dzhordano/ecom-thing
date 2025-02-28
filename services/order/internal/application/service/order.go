package service

import (
	"context"
	"errors"
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain/repository"
	"github.com/dzhordano/ecom-thing/services/order/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrderService struct {
	log  logger.BaseLogger
	repo repository.OrderRepository
}

func NewOrderService(log logger.BaseLogger, repo repository.OrderRepository) interfaces.OrderService {
	return &OrderService{
		log:  log,
		repo: repo,
	}
}

// CreateOrder implements interfaces.OrderService.
func (o *OrderService) CreateOrder(ctx context.Context, info dto.CreateOrderRequest) (*domain.Order, error) {
	var disc *domain.Coupon
	var err error

	if info.Coupon == "" {
		disc = &domain.Coupon{
			Discount: 0,
		}
	} else {
		disc, err = o.repo.GetCoupon(ctx, info.Coupon)
		if err != nil {
			return nil, err
		}

		if disc.ValidTo.Before(time.Now()) {
			return nil, domain.ErrCouponExpired
		}
	}

	order, err := domain.NewOrder(
		uuid.New(), // FIXME Щас рандомный пользотель. Потом получать из контекста.
		info.Description,
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

	if err := o.repo.Save(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// GetById implements interfaces.OrderService.
func (o *OrderService) GetById(ctx context.Context, orderId uuid.UUID) (*domain.Order, error) {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		return nil, err
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	return order, nil
}

// ListByUser implements interfaces.OrderService.
func (o *OrderService) ListByUser(ctx context.Context, limit uint64, offset uint64) ([]*domain.Order, error) {
	// FIXME Щас тут рандомный uuid, потом из контекста.

	randUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return o.repo.ListByUser(ctx, randUUID.String())
}

// Search implements interfaces.OrderService.
func (o *OrderService) SearchOrders(ctx context.Context, filters map[string]any) ([]*domain.Order, error) {
	params := domain.NewSearchParams(filters)

	if err := params.Validate(); err != nil {
		o.log.Error("failed to search orders", zap.Error(err))
		return nil, err
	}

	orders, err := o.repo.Search(ctx, params)
	if err != nil {
		o.log.Error("failed to search orders", zap.Error(err))
		return nil, errors.Unwrap(err)
	}

	o.log.Debug("orders retrieved", zap.Int("count", len(orders)))

	return orders, nil
}

// UpdateOrder implements interfaces.OrderService.
func (o *OrderService) UpdateOrder(ctx context.Context, info dto.UpdateOrderRequest) (*domain.Order, error) {
	order, err := o.repo.GetById(ctx, info.OrderID.String())
	if err != nil {
		return nil, err
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if info.Description != nil {
		order.Description = *info.Description
	}

	if info.Status != nil {
		s, err := domain.NewStatus(*info.Status)
		if err != nil {
			return nil, err
		}
		order.Status = s
	}

	if info.TotalPrice != nil {
		order.TotalPrice = *info.TotalPrice
	}

	if info.PaymentMethod != nil {
		pm, err := domain.NewPaymentMethod(*info.PaymentMethod)
		if err != nil {
			return nil, err
		}
		order.PaymentMethod = pm
	}

	if info.DeliveryMethod != nil {
		dm, err := domain.NewDeliveryMethod(*info.DeliveryMethod)
		if err != nil {
			return nil, err
		}
		order.DeliveryMethod = dm
	}

	if info.DeliveryAddress != nil {
		order.DeliveryAddress = *info.DeliveryAddress
	}

	if !info.DeliveryDate.IsZero() {
		order.DeliveryDate = info.DeliveryDate
	}

	if len(info.Items) > 0 {
		order.Items = info.Items
	}

	// Допроверить поля на валидность.
	if err = order.Validate(); err != nil {
		return nil, err
	}

	if err := o.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// DeleteOrder implements interfaces.OrderService.
func (o *OrderService) DeleteOrder(ctx context.Context, orderId uuid.UUID) error {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		return err
	}

	// Чтобы компилятор не жаловался...
	if order.ID == uuid.Nil {
		return domain.ErrOrderNotFound
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if err := o.repo.Delete(ctx, orderId.String()); err != nil {
		return err
	}

	return nil
}

// CompleteOrder implements interfaces.OrderService.
func (o *OrderService) CompleteOrder(ctx context.Context, orderId uuid.UUID) error {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		return err
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if order.Status == domain.OrderCancelled {
		return domain.ErrOrderAlreadyCancelled
	}

	order.Status = domain.OrderCompleted
	order.UpdatedAt = time.Now()

	if err := o.repo.Update(ctx, order); err != nil {
		return err
	}

	return nil
}

// CancelOrder implements interfaces.OrderService.
func (o *OrderService) CancelOrder(ctx context.Context, orderId uuid.UUID) error {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		return err
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if order.Status == domain.OrderCompleted {
		return domain.ErrOrderAlreadyCompleted
	}

	order.Status = domain.OrderCancelled
	order.UpdatedAt = time.Now()

	if err := o.repo.Update(ctx, order); err != nil {
		return err
	}

	return nil
}
