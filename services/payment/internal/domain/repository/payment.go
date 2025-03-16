package repository

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
)

type PaymentRepository interface {
	Save(ctx context.Context, payment *domain.Payment) error
	GetById(ctx context.Context, paymentId, userId string) (*domain.Payment, error)
	ListByUser(ctx context.Context, userId string, limit, offset uint64) ([]*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) error
	Delete(ctx context.Context, paymentId string) error
}
