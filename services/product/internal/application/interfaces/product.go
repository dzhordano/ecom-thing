package interfaces

import (
	"context"
	"github.com/google/uuid"

	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
)

type ProductService interface {
	CreateProduct(ctx context.Context, name, description string, price float64, quantity uint64) (*domain.Product, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetAllProducts(ctx context.Context) ([]*domain.Product, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, name, description string, price float64, quantity uint64) (*domain.Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error)
}
