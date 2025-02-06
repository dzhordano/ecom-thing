package repository

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"github.com/google/uuid"
)

type ProductRepository interface {
	Save(ctx context.Context, product *domain.Product) error
	Update(ctx context.Context, product *domain.Product) error
	Deactivate(ctx context.Context, id uuid.UUID) error

	GetById(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	Search(ctx context.Context, options domain.SearchOptions, limit, offset uint64) ([]*domain.Product, error)
}
