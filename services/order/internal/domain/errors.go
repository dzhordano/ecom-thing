package domain

// Just for convienience.

import "errors"

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrInvalidOrderStatus    = errors.New("invalid order status")
	ErrInvalidCurrency       = errors.New("invalid currency")
	ErrInvalidPaymentMethod  = errors.New("invalid payment method")
	ErrInvalidDeliveryMethod = errors.New("invalid delivery method")
	ErrInvalidArgument       = errors.New("invalid argument")
	ErrInvalidDescription    = errors.New("invalid description")

	ErrInternal               = errors.New("internal error")
	ErrInvalidUUID            = errors.New("invalid uuid")
	ErrInvalidPrice           = errors.New("invalid price")
	ErrInvalidDiscount        = errors.New("invalid discount")
	ErrInvalidDeliveryAddress = errors.New("invalid delivery address")
	ErrInvalidDeliveryDate    = errors.New("invalid delivery date")
	ErrInvalidOrderItems      = errors.New("invalid order items")

	ErrOrderAlreadyCompleted = errors.New("order already completed")
	ErrOrderAlreadyCancelled = errors.New("order already cancelled")

	ErrCouponExpired   = errors.New("coupon expired")
	ErrCouponNotFound  = errors.New("coupon not found")
	ErrCouponNotActive = errors.New("coupon not active")

	ErrProductUnavailable   = errors.New("product unavailable")
	ErrInventoryUnavailable = errors.New("inventory unavailable")
)

func CheckIfCriticalError(err error) bool {
	return !(errors.Is(err, ErrOrderNotFound) ||
		errors.Is(err, ErrInvalidOrderStatus) ||
		errors.Is(err, ErrInvalidCurrency) ||
		errors.Is(err, ErrInvalidPaymentMethod) ||
		errors.Is(err, ErrInvalidDeliveryMethod) ||
		errors.Is(err, ErrInvalidArgument) ||
		errors.Is(err, ErrInvalidDescription) ||
		errors.Is(err, ErrInvalidUUID) ||
		errors.Is(err, ErrInvalidPrice) ||
		errors.Is(err, ErrInvalidDiscount) ||
		errors.Is(err, ErrInvalidDeliveryAddress) ||
		errors.Is(err, ErrInvalidDeliveryDate) ||
		errors.Is(err, ErrInvalidOrderItems) ||
		errors.Is(err, ErrOrderAlreadyCompleted) ||
		errors.Is(err, ErrOrderAlreadyCancelled) ||
		errors.Is(err, ErrCouponExpired) ||
		errors.Is(err, ErrCouponNotFound) ||
		errors.Is(err, ErrCouponNotActive))
}
