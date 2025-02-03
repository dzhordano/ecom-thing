package repository

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"github.com/google/uuid"
)

type ProductRepository interface {
	Save(ctx context.Context, product *domain.Product) error
	Get(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetAll(ctx context.Context) ([]*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}
