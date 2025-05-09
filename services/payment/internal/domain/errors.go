package domain

import (
	"errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrInvalidArgument         = errors.New("invalid argument")
	ErrInvalidStatus           = errors.New("invalid status")
	ErrInvalidCurrency         = errors.New("invalid currency")
	ErrInvalidPaymentMethod    = errors.New("invalid payment method")
	ErrInvalidPayment          = errors.New("invalid payment") // Means that payment has invalid status for example
	ErrPaymentAlreadyPending   = errors.New("payment already pending")
	ErrPaymentAlreadyCompleted = errors.New("payment already completed")
	ErrPaymentAlreadyExists    = errors.New("payment already exists") // Payment for a certain order is already created.
	ErrPaymentNotFound         = errors.New("payment not found")

	// Critical ones --->
	ErrPaymentCancelled = errors.New("payment cancelled") // FIXME надо ли?
	ErrPaymentFailed    = errors.New("payment failed")
	ErrInternal         = errors.New("internal error")
)

var CriticalErrors = map[error]struct{}{
	ErrPaymentCancelled: {},
	ErrPaymentFailed:    {},
	ErrInternal:         {},
}

// If errors are critical, stacktrace will be included in logs.
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
	case errors.Is(e.Code, ErrInvalidArgument):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrPaymentAlreadyPending):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrPaymentAlreadyCompleted):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrPaymentAlreadyExists):
		return codes.AlreadyExists
	case errors.Is(e.Code, ErrPaymentNotFound):
		return codes.NotFound
	case errors.Is(e.Code, ErrPaymentCancelled):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrPaymentFailed):
		return codes.InvalidArgument
	default:
		return codes.Internal
	}
}
