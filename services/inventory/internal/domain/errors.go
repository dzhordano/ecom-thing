package domain

import "errors"

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
