package domain

// Just for convenience.

import (
	"errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrInvalidOrderStatus    = errors.New("invalid order status")
	ErrInvalidCurrency       = errors.New("invalid currency")
	ErrInvalidPaymentMethod  = errors.New("invalid payment method")
	ErrInvalidDeliveryMethod = errors.New("invalid delivery method")
	ErrInvalidArgument       = errors.New("invalid argument")
	ErrInvalidDescription    = errors.New("invalid description")

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

	ErrNotEnoughQuantity = errors.New("not enough quantity")

	ErrProductUnavailable   = errors.New("product unavailable")
	ErrInventoryUnavailable = errors.New("inventory unavailable")
)

var CriticalErrors = map[error]struct{}{
	ErrProductUnavailable:   {},
	ErrInventoryUnavailable: {},
}

func CheckIfCriticalError(err error) bool {
	_, ok := CriticalErrors[err]
	return ok
}

type AppError struct {
	Code error
	Msg  string
}

func NewAppError(code error, message string) *AppError {
	return &AppError{
		Code: code,
		Msg:  message,
	}
}

func (e *AppError) Error() string {
	return e.Msg
}

func (e *AppError) Unwrap() error {
	return e.Code
}

func (e *AppError) Is(target error) bool {
	return errors.Is(e.Code, target)
}

func (e *AppError) GRPCCode() codes.Code {
	switch {
	case errors.Is(e.Code, ErrOrderNotFound):
		return codes.NotFound
	case errors.Is(e.Code, ErrInvalidArgument):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrInvalidUUID):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrOrderAlreadyCompleted):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrOrderAlreadyCancelled):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrCouponExpired):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrCouponNotActive):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrCouponNotFound):
		return codes.NotFound
	case errors.Is(e.Code, ErrNotEnoughQuantity):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrProductUnavailable):
		return codes.NotFound
	case errors.Is(e.Code, ErrInventoryUnavailable):
		return codes.NotFound
	default:
		return codes.Internal
	}
}
