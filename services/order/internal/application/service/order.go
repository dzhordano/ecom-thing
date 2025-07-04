package service

import (
	"context"
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain/repository"
	"github.com/dzhordano/ecom-thing/services/order/pkg/logger"
	"github.com/google/uuid"
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
	log              logger.Logger
	productService   interfaces.ProductService
	inventoryService interfaces.InventoryService
	repo             repository.OrderRepository
}

func NewOrderService(l logger.Logger, ps interfaces.ProductService, is interfaces.InventoryService, r repository.OrderRepository) interfaces.OrderService {
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
			o.log.Error("failed to get coupon", "error", err, "coupon", info.Coupon)
			return nil, domain.NewAppError(err, "failed to get coupon")
		}

		// Если купон просрочен - ошибка.
		if disc.ValidTo.Before(time.Now()) {
			o.log.Error("failed to get coupon", "error", domain.ErrCouponExpired, "coupon", info.Coupon)
			return nil, domain.NewAppError(domain.ErrCouponExpired, "coupon is expired")
		}

		// Купон есть, но не активен.
		if disc.ValidFrom.After(time.Now()) {
			o.log.Error("failed to get coupon", "error", domain.ErrCouponNotActive, "coupon", info.Coupon)
			return nil, domain.NewAppError(domain.ErrCouponNotActive, "coupon is not active")
		}
	}

	timeout, cancel := context.WithTimeout(ctx, 5*time.Second) // FIXME Тоже хардкод
	defer cancel()

	var totalPrice float64
	for _, item := range info.Items {
		price, isActive, err := o.productService.GetProductInfo(timeout, item.ProductID)
		if err != nil {
			o.log.Error("failed to get product info", "error", err, "product_id", item.ProductID)
			return nil, domain.NewAppError(err, "failed to get product info")
		}

		if !isActive {
			o.log.Error("product unavailable", "error", domain.ErrProductUnavailable, "product_id", item.ProductID)
			return nil, domain.NewAppError(domain.ErrProductUnavailable, "product unavailable")
		}

		totalPrice += float64(item.Quantity) * price
	}

	order, err := domain.NewOrder(
		uuid.New(), // FIXME Щас рандомный пользователь. Потом получать из контекста.
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
		o.log.Error("failed to create order", "error", err)
		return nil, domain.NewAppError(err, err.Error()) // TODO В идеале все таки валидацию вне NewOrder, т.к. там может вернуть что uuid не получилос сгенерить.
	}

	items := make(map[string]uint64)
	for _, item := range order.Items {
		items[item.ProductID.String()] += item.Quantity
	}

	timeout, cancel = context.WithTimeout(ctx, 5*time.Second) // FIXME Тоже хардкод
	defer cancel()

	isReservable, err := o.inventoryService.IsReservable(timeout, items)
	if err != nil {
		o.log.Error("failed to check if items reservable", "error", err)
		return nil, domain.NewAppError(err, "failed to check if items reservable")
	}

	if !isReservable {
		o.log.Error("failed to reserve order", "error", domain.ErrNotEnoughQuantity)
		return nil, domain.NewAppError(domain.ErrNotEnoughQuantity, "not enough quantity")
	}

	if err = o.repo.Save(ctx, order); err != nil {
		o.log.Error("failed to save order", "error", err)
		return nil, domain.NewAppError(err, "failed to save order")
	}

	o.log.Debug("order created", "order_id", order.ID.String())

	return order, nil
}

// GetById implements interfaces.OrderService.
func (o *OrderService) GetById(ctx context.Context, orderId uuid.UUID) (*domain.Order, error) {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		o.log.Error("failed to get order", "error", err, "order_id", orderId.String())
		return nil, domain.NewAppError(err, "failed to get order")
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	o.log.Debug("order retrieved", "order_id", order.ID.String())

	return order, nil
}

// ListByUser implements interfaces.OrderService.
func (o *OrderService) ListByUser(ctx context.Context, limit uint64, offset uint64) ([]*domain.Order, error) {
	// FIXME Щас тут рандомный uuid, потом из контекста.

	randUUID, err := uuid.NewRandom()
	if err != nil {
		o.log.Error("failed to list orders", "error", err)
		return nil, domain.NewAppError(err, "failed to list orders")
	}

	orders, err := o.repo.ListByUser(ctx, randUUID.String())
	if err != nil {
		o.log.Error("failed to list orders", "error", err, "user_id", randUUID.String())
		return nil, domain.NewAppError(err, "failed to list orders")
	}

	o.log.Debug("orders retrieved", "count", len(orders))

	return orders, nil
}

// Search implements interfaces.OrderService.
func (o *OrderService) SearchOrders(ctx context.Context, filters map[string]any) ([]*domain.Order, error) {
	params := domain.NewSearchParams(filters)

	if err := params.Validate(); err != nil {
		o.log.Error("failed to search orders", "error", err)
		return nil, domain.NewAppError(err, err.Error())
	}

	orders, err := o.repo.Search(ctx, params)
	if err != nil {
		o.log.Error("failed to search orders", "error", err)
		return nil, domain.NewAppError(err, "failed to search orders")
	}

	o.log.Debug("orders retrieved", "count", len(orders))

	return orders, nil
}

// FIXME тут возможен возврат одной ошибки, когда можно вернуть несколько... испрвить.
// UpdateOrder implements interfaces.OrderService.
func (o *OrderService) UpdateOrder(ctx context.Context, info dto.UpdateOrderRequest) (*domain.Order, error) {
	order, err := o.repo.GetById(ctx, info.OrderID.String())
	if err != nil {
		o.log.Error("failed to update order", "error", err, "order_id", info.OrderID.String())
		return nil, domain.NewAppError(err, "failed to get order")
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if info.Description != nil {
		order.Description = *info.Description
	}

	if info.Status != nil {
		order.Status = domain.Status(*info.Status)
	}

	if info.TotalPrice != nil {
		order.TotalPrice = *info.TotalPrice
	}

	if info.PaymentMethod != nil {
		order.PaymentMethod = domain.PaymentMethod(*info.PaymentMethod)
	}

	if info.DeliveryMethod != nil {
		order.DeliveryMethod = domain.DeliveryMethod(*info.DeliveryMethod)
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

	if err = order.Validate(); err != nil {
		o.log.Error("failed to update order", "error", err, "order_id", info.OrderID.String())
		return nil, domain.NewAppError(err, err.Error())
	}

	if err := o.repo.Update(ctx, order); err != nil {
		o.log.Error("failed to update order", "error", err, "order_id", info.OrderID.String())
		return nil, domain.NewAppError(err, "failed to update order")
	}

	o.log.Debug("order updated", "order_id", order.ID.String())

	return order, nil
}

// DeleteOrder implements interfaces.OrderService.
func (o *OrderService) DeleteOrder(ctx context.Context, orderId uuid.UUID) error {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		o.log.Error("failed to delete order", "error", err, "order_id", orderId.String())
		return domain.NewAppError(err, "failed to get order")
	}

	// Чтобы компилятор не жаловался...
	// FIXME не забыть убрать
	if order.ID == uuid.Nil {
		return domain.NewAppError(domain.ErrOrderNotFound, "order not found")
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if err := o.repo.Delete(ctx, orderId.String()); err != nil {
		o.log.Error("failed to delete order", "error", err, "order_id", orderId.String())
		return domain.NewAppError(err, "failed to delete order")
	}

	o.log.Debug("order deleted", "order_id", order.ID.String())

	return nil
}

// CompleteOrder implements interfaces.OrderService.
func (o *OrderService) CompleteOrder(ctx context.Context, orderId uuid.UUID) error {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		o.log.Error("failed to complete order", "error", err, "order_id", orderId.String())
		return domain.NewAppError(err, "failed to get order")
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if order.Status == domain.OrderCancelled {
		o.log.Error("failed to complete order", "error", domain.ErrOrderAlreadyCancelled, "order_id", orderId.String())
		return domain.NewAppError(domain.ErrOrderAlreadyCancelled, "order already cancelled")
	}

	order.Status = domain.OrderCompleted
	order.UpdatedAt = time.Now()

	// for _, item := range order.Items {
	// 	if err := o.inventoryService.SubReservedQuantity(ctx, item.ProductID, item.Quantity); err != nil {
	// 		o.log.Error("failed to complete order", "error", err)
	// 		return err
	// 	}
	// }

	items := map[string]uint64{}
	for _, item := range order.Items {
		items[item.ProductID.String()] = item.Quantity
	}

	if err := o.repo.Update(ctx, order); err != nil {
		o.log.Error("failed to complete order", "error", err, "order_id", orderId.String())
		return domain.NewAppError(err, "failed to complete order")
	}

	o.log.Debug("order completed", "order_id", order.ID.String())

	return nil
}

// CancelOrder implements interfaces.OrderService.
func (o *OrderService) CancelOrder(ctx context.Context, orderId uuid.UUID) error {
	order, err := o.repo.GetById(ctx, orderId.String())
	if err != nil {
		o.log.Error("failed to cancel order", "error", err, "order_id", orderId.String())
		return domain.NewAppError(err, "failed to get order")
	}

	// FIXME Тут проверка на принадлежность пользователю. Получение Id пользователя из контекста.

	if order.Status == domain.OrderCompleted {
		o.log.Error("failed to cancel order", "error", domain.ErrOrderAlreadyCompleted, "order_id", orderId.String())
		return domain.NewAppError(domain.ErrOrderAlreadyCompleted, "order already completed")
	}

	order.Status = domain.OrderCancelled
	order.UpdatedAt = time.Now()

	items := map[string]uint64{}
	for _, item := range order.Items {
		items[item.ProductID.String()] = item.Quantity
	}

	if err := o.repo.Update(ctx, order); err != nil {
		o.log.Error("failed to cancel order", "error", err, "order_id", orderId.String())
		return domain.NewAppError(err, "failed to cancel order")
	}

	o.log.Debug("order cancelled", "order_id", order.ID.String())

	return nil
}
