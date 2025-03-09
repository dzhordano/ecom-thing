package domain

import "errors"

var (
	ErrInvalidArgument      = errors.New("invalid argument")
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
)
