package interfaces

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/payment/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
	"github.com/google/uuid"
)

type PaymentService interface {
	CreatePayment(ctx context.Context, req dto.CreatePaymentRequest) (*domain.Payment, error)
	GetPaymentStatus(ctx context.Context, paymentId, userId uuid.UUID) (string, error)
	RetryPayment(ctx context.Context, paymentId, userId uuid.UUID) error
	// TODO Тут понять юзкейсы
	CancelPayment(ctx context.Context, orderId, userId uuid.UUID) error
	ConfirmPayment(ctx context.Context, orderId, userId uuid.UUID) error
}
