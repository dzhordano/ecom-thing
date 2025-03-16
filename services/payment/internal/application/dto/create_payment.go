package dto

import "github.com/google/uuid"

type CreatePaymentRequest struct {
	OrderId    uuid.UUID
	UserId     uuid.UUID
	Currency   string
	TotalPrice float64

	PaymentMethod string
	Description   string
	RedirectURL   string
}
