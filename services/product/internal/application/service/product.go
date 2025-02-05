package service

import (
	"context"
	"errors"
	"github.com/dzhordano/ecom-thing/services/product/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain/repository"
	"github.com/google/uuid"
	"log/slog"
)

type ProductService struct {
	log  *slog.Logger
	repo repository.ProductRepository
}

func NewProductService(log *slog.Logger, repo repository.ProductRepository) interfaces.ProductService {
	return &ProductService{
		log:  log,
		repo: repo,
	}
}

func (p *ProductService) CreateProduct(ctx context.Context, name, description, category string, price float64) (*domain.Product, error) {
	userId, err := uuid.NewUUID()
	if err != nil {
		p.log.Error("failed to create product", slog.String("error", err.Error()))
		return nil, err
	}

	product, err := domain.NewValidatedProduct(userId, name, description, category, price)
	if err != nil {
		p.log.Error("failed to create product", slog.String("error", err.Error()))
		return nil, err
	}

	if err := p.repo.Save(ctx, product); err != nil {
		p.log.Error("failed to save product", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("product created", "id", product.ID)

	return product, nil
}

func (p *ProductService) GetProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := p.repo.Get(ctx, id)
	if err != nil {
		p.log.Error("failed to get product", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("product retrieved", "id", product.ID)

	return product, nil
}

func (p *ProductService) GetAllProducts(ctx context.Context) ([]*domain.Product, error) {
	products, err := p.repo.GetAll(ctx)
	if err != nil {
		p.log.Error("failed to get all products", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("all products retrieved")

	return products, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, name, description, category string, price float64) (*domain.Product, error) {
	product, err := p.repo.Get(ctx, id)
	if err != nil {
		p.log.Error("failed to update product", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	product.Update(name, description, category, price)

	if err := product.Validate(); err != nil {
		p.log.Error("failed to update product", slog.String("error", err.Error()))
		return nil, err
	}

	if err := p.repo.Update(ctx, product); err != nil {
		p.log.Error("failed to update product", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("product updated", "id", product.ID)

	return product, nil
}

func (p *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := p.repo.Get(ctx, id)
	if err != nil {
		p.log.Error("failed to delete product", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	if err := p.repo.Delete(ctx, id); err != nil {
		p.log.Error("failed to delete product", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("product deleted", "id", id)

	return product, nil
}
