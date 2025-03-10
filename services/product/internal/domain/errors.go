package domain

import "errors"

var (
	ErrInvalidArgument      = errors.New("invalid argument")
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
)

// There is not critical ones, so if these aren't preset -> give stacktrace
func CheckIfCriticalError(err error) bool {
	return !(errors.Is(err, ErrInvalidArgument) || errors.Is(err, ErrProductNotFound) || errors.Is(err, ErrProductAlreadyExists))
}
