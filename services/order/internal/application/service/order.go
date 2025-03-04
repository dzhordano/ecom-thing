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

const (
	// Operations used in inventory service here. Idk maybe fix, looks bad enough.
	OperationAdd       = "add"
	OperationSub       = "sub"
	OperationLock      = "lock"
	OperationUnlock    = "unlock"
	OperationSubLocked = "sub_locked"
)

type OrderService struct {
	log              logger.BaseLogger
	productService   interfaces.ProductService
	inventoryService interfaces.InventoryService
	repo             repository.OrderRepository
}

func NewOrderService(l logger.BaseLogger, ps interfaces.ProductService, is interfaces.InventoryService, r repository.OrderRepository) interfaces.OrderService {
	return &OrderService{
		log:              l,
		productService:   ps,
		inventoryService: is,
		repo:             r,
	}
}

// CreateOrder implements interfaces.OrderService.
func (o *OrderService) CreateOrder(ctx context.Context, info dto.CreateOrderRequest) (*domain.Order, error) {
	disc := &domain.Coupon{}
	var err error

	if info.Coupon != "" {
		disc, err = o.repo.GetCoupon(ctx, info.Coupon)
		if err != nil {
			o.log.Error("failed to get coupon", zap.Error(err))
			return nil, err
		}

		// Если купон просрочен - ошибка.
		if disc.ValidTo.Before(time.Now()) {
			o.log.Error("failed to get coupon", zap.Error(domain.ErrCouponExpired))
			return nil, domain.ErrCouponExpired
		}

		// Купон есть, но не активен.
		if disc.ValidFrom.After(time.Now()) {
			o.log.Error("failed to get coupon", zap.Error(domain.ErrCouponNotActive))
			return nil, domain.ErrCouponNotActive
		}
	}

	var totalPrice float64
	for _, item := range info.Items {
		price, isActive, err := o.productService.GetProductInfo(ctx, item.ProductID)
		if err != nil {
			o.log.Error("failed to get product info", zap.Error(err))
			return nil, err
		}

		if !isActive {
			o.log.Error("failed to get product info", zap.Error(domain.ErrProductUnavailable))
			return nil, domain.ErrProductUnavailable
		}

		totalPrice += float64(item.Quantity) * price
	}

	order, err := domain.NewOrder(
		uuid.New(), // FIXME Щас рандомный пользотель. Потом получать из контекста.
		info.Description,
		domain.OrderPending.String(),
		info.Currency,
		totalPrice,
		disc.Discount,
		info.PaymentMethod,
		info.DeliveryMethod,
		info.DeliveryAddress,
		info.DeliveryDate,
		info.Items,
	)
	if err != nil {
		o.log.Error("failed to create order", zap.Error(err))
		return nil, domain.ErrInternal // Internal так как не хочу давать контекста туда куда-то.
	}

	items := map[string]uint64{}
	for _, item := range info.Items {
		items[item.ProductID.String()] = item.Quantity
	}

	// FIXME THIS ADDS TON OF INCOSISTENCY.
	//
	// Сделать через outbox. (pub: inventory-events -> reservation-request) в outbox.go
	if err := o.inventoryService.SetItemsWithOp(ctx, items, OperationLock); err != nil {
		o.log.Error("failed to reserve product", zap.Error(err))
		return nil, err
	}

	if err = o.repo.Save(ctx, order); err != nil {
		o.log.Error("failed to save order", zap.Error(err))

		if err := o.inventoryService.SetItemsWithOp(ctx, items, OperationUnlock); err != nil {
			o.log.Error("failed to remove reserved product", zap.Error(err))
		}

		return nil, err
	}

	o.log.Debug("order created", zap.String("order_id", order.ID.String()))

	return order, nil
}

// GetById implements interfaces.OrderService.
func (o *OrderService) GetById(ctx context.Context, orderId uuid.UUID) (*domain.Order, error) {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		o.log.Error("failed to get order", zap.Error(err))
		return nil, err
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	o.log.Debug("order retrieved", zap.String("order_id", order.ID.String()))

	return order, nil
}

// ListByUser implements interfaces.OrderService.
func (o *OrderService) ListByUser(ctx context.Context, limit uint64, offset uint64) ([]*domain.Order, error) {
	// FIXME Щас тут рандомный uuid, потом из контекста.

	randUUID, err := uuid.NewRandom()
	if err != nil {
		o.log.Error("failed to list orders", zap.Error(err))
		return nil, err
	}

	orders, err := o.repo.ListByUser(ctx, randUUID.String())
	if err != nil {
		o.log.Error("failed to list orders", zap.Error(err))
		return nil, err
	}

	o.log.Debug("orders retrieved", zap.Int("count", len(orders)))

	return orders, nil
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
		o.log.Error("failed to update order", zap.Error(err))
		return nil, err
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if info.Description != nil {
		order.Description = *info.Description
	}

	if info.Status != nil {
		s, err := domain.NewStatus(*info.Status)
		if err != nil {
			o.log.Error("failed to update order", zap.Error(err))
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
			o.log.Error("failed to update order", zap.Error(err))
			return nil, err
		}
		order.PaymentMethod = pm
	}

	if info.DeliveryMethod != nil {
		dm, err := domain.NewDeliveryMethod(*info.DeliveryMethod)
		if err != nil {
			o.log.Error("failed to update order", zap.Error(err))
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
		o.log.Error("failed to update order", zap.Error(err))
		return nil, err
	}

	if err := o.repo.Update(ctx, order); err != nil {
		o.log.Error("failed to update order", zap.Error(err))
		return nil, err
	}

	o.log.Debug("order updated", zap.String("order_id", order.ID.String()))

	return order, nil
}

// DeleteOrder implements interfaces.OrderService.
func (o *OrderService) DeleteOrder(ctx context.Context, orderId uuid.UUID) error {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		o.log.Error("failed to delete order", zap.Error(err))
		return err
	}

	// Чтобы компилятор не жаловался...
	// FIXME не забыть убрать
	if order.ID == uuid.Nil {
		return domain.ErrOrderNotFound
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if err := o.repo.Delete(ctx, orderId.String()); err != nil {
		o.log.Error("failed to delete order", zap.Error(err))
		return err
	}

	o.log.Debug("order deleted", zap.String("order_id", order.ID.String()))

	return nil
}

// CompleteOrder implements interfaces.OrderService.
func (o *OrderService) CompleteOrder(ctx context.Context, orderId uuid.UUID) error {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		o.log.Error("failed to complete order", zap.Error(err))
		return err
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if order.Status == domain.OrderCancelled {
		o.log.Error("failed to complete order", zap.Error(domain.ErrOrderAlreadyCancelled))
		return domain.ErrOrderAlreadyCancelled
	}

	order.Status = domain.OrderCompleted
	order.UpdatedAt = time.Now()

	// for _, item := range order.Items {
	// 	if err := o.inventoryService.SubReservedQuantity(ctx, item.ProductID, item.Quantity); err != nil {
	// 		o.log.Error("failed to complete order", zap.Error(err))
	// 		return err
	// 	}
	// }

	items := map[string]uint64{}
	for _, item := range order.Items {
		items[item.ProductID.String()] = item.Quantity
	}

	if err = o.inventoryService.SetItemsWithOp(ctx, items, OperationSubLocked); err != nil {
		o.log.Error("failed to complete order", zap.Error(err))
		return err
	}

	if err := o.repo.Update(ctx, order); err != nil {
		o.log.Error("failed to complete order", zap.Error(err))
		return err
	}

	o.log.Debug("order completed", zap.String("order_id", order.ID.String()))

	return nil
}

// CancelOrder implements interfaces.OrderService.
func (o *OrderService) CancelOrder(ctx context.Context, orderId uuid.UUID) error {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		o.log.Error("failed to cancel order", zap.Error(err))
		return err
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if order.Status == domain.OrderCompleted {
		o.log.Error("failed to cancel order", zap.Error(domain.ErrOrderAlreadyCompleted))
		return domain.ErrOrderAlreadyCompleted
	}

	order.Status = domain.OrderCancelled
	order.UpdatedAt = time.Now()

	items := map[string]uint64{}
	for _, item := range order.Items {
		items[item.ProductID.String()] = item.Quantity
	}

	if err = o.inventoryService.SetItemsWithOp(ctx, items, OperationUnlock); err != nil {
		o.log.Error("failed to complete order", zap.Error(err))
		return err
	}
	if err := o.repo.Update(ctx, order); err != nil {
		o.log.Error("failed to cancel order", zap.Error(err))
		return err
	}

	o.log.Debug("order cancelled", zap.String("order_id", order.ID.String()))

	return nil
}
