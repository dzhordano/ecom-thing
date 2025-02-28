package interfaces

import (
	"context"

	"github.com/google/uuid"

	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
)

type ProductService interface {
	CreateProduct(ctx context.Context, name, description, category string, price float64) (*domain.Product, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, name, description, category string, isActive bool, price float64) (*domain.Product, error)
	DeactivateProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error)

	GetById(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	SearchProducts(ctx context.Context, filters map[string]any) ([]*domain.Product, error)
}
