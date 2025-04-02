package domain

import "errors"

var (
	ErrInvalidArgument      = errors.New("invalid argument")
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
)

var CriticalErrors = map[error]struct{}{}

func CheckIfCriticalError(err error) bool {
	_, ok := CriticalErrors[err]
	return ok
}
