package domain

import (
	"errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrOperationUnknown  = errors.New("operation unknown")
	ErrProductNotFound   = errors.New("product not found")
	ErrNotEnoughQuantity = errors.New("not enough quantity")
)

var CriticalErrors = map[error]struct{}{}

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

func (e *AppError) Is(target error) bool {
	return errors.Is(e.Code, target)
}

func (e *AppError) Unwrap() error {
	return e.Code
}

func (e *AppError) GRPCCode() codes.Code {
	switch {
	case errors.Is(e.Code, ErrOperationUnknown):
		return codes.InvalidArgument
	case errors.Is(e.Code, ErrProductNotFound):
		return codes.NotFound
	case errors.Is(e.Code, ErrOperationUnknown):
		return codes.InvalidArgument
	default:
		return codes.Internal
	}
}

// var (
// 	ErrOperationUnknown  = errors.New("operation unknown")
// 	ErrProductNotFound   = errors.New("product not found")
// 	ErrNotEnoughQuantity = errors.New("not enough quantity")
// )
