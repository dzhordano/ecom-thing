package interfaces

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/payment/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
	"github.com/google/uuid"
)

type Billing interface {
	// NewPayment handles process of payment. Is BLOCKING (supposedly) operation.
	NewPayment(ctx context.Context, currency string, totalPrice float64, paymentDescription string) error
}

type PaymentService interface {
	CreatePayment(ctx context.Context, req dto.CreatePaymentRequest) (*domain.Payment, error)
	GetPaymentStatus(ctx context.Context, paymentId, userId uuid.UUID) (string, error)
	RetryPayment(ctx context.Context, paymentId, userId uuid.UUID) error
	// TODO Тут понять юзкейсы
	CancelPayment(ctx context.Context, paymentId, userId uuid.UUID) error
	ConfirmPayment(ctx context.Context, paymentId, userId uuid.UUID) error
}
