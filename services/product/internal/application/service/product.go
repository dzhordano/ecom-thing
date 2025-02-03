package service

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain/repository"
	"github.com/google/uuid"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) interfaces.ProductService {
	return &ProductService{
		repo: repo,
	}
}

func (p ProductService) CreateProduct(ctx context.Context, name, description string, price float64, quantity uint64) (*domain.Product, error) {
	// TODO implement me
	panic("implement me")
}

func (p ProductService) GetProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	// TODO implement me
	panic("implement me")
}

func (p ProductService) GetAllProducts(ctx context.Context) ([]*domain.Product, error) {
	// TODO implement me
	panic("implement me")
}

func (p ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, name, description string, price float64, quantity uint64) (*domain.Product, error) {
	// TODO implement me
	panic("implement me")
}

func (p ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	// TODO implement me
	panic("implement me")
}
