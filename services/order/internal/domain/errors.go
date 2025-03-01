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

	ErrProductUnavailable = errors.New("product unavailable")
)
