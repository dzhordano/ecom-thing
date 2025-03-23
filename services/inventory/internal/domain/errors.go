package domain

import "errors"

var (
	ErrOperationUnknown  = errors.New("operation unknown")
	ErrProductNotFound   = errors.New("product not found")
	ErrNotEnoughQuantity = errors.New("not enough quantity")
)

func CheckIfCriticalError(err error) bool {
	// No particular critical errors, so mark everything as critical expect the ones we know.
	return !(errors.Is(err, ErrNotEnoughQuantity) || errors.Is(err, ErrProductNotFound) || errors.Is(err, ErrOperationUnknown))
}
