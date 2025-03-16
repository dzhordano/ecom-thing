package billing

import "context"

// Service for payment
type Billing interface {
	CreatePayment(ctx context.Context, currency, amount, redirectURL, paymentData string) error
	CancelPayment(ctx context.Context, paymentId string) error
}
