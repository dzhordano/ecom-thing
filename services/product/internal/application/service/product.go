package service

import (
	"context"
	"errors"

	"github.com/dzhordano/ecom-thing/services/product/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain/repository"
	"github.com/dzhordano/ecom-thing/services/product/pkg/logger"
	"github.com/google/uuid"
)

type ProductService struct {
	log  logger.Logger
	repo repository.ProductRepository
}

func NewProductService(log logger.Logger, repo repository.ProductRepository) interfaces.ProductService {
	return &ProductService{
		log:  log,
		repo: repo,
	}
}

func (p *ProductService) CreateProduct(ctx context.Context, name, description, category string, price float64) (*domain.Product, error) {
	userId, err := uuid.NewUUID()
	if err != nil {
		p.log.Error("failed to create product", "error", err)
		return nil, err
	}

	product, err := domain.NewValidatedProduct(userId, name, description, category, price)
	if err != nil {
		p.log.Error("failed to create product", "error", err)
		return nil, err
	}

	if err := p.repo.Save(ctx, product); err != nil {
		p.log.Error("failed to save product", "error", err)
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("product created", "id", product.ID.String())

	return product, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, name, description, category string, isActive bool, price float64) (*domain.Product, error) {
	product, err := p.repo.GetById(ctx, id)
	if err != nil {
		p.log.Error("failed to update product", "error", err, "product_id", id.String())
		return nil, errors.Unwrap(err)
	}

	product.Update(name, description, category, isActive, price)

	if err := product.Validate(); err != nil {
		p.log.Error("failed to update product", "error", err, "product_id", id.String())
		return nil, err
	}

	if err := p.repo.Update(ctx, product); err != nil {
		p.log.Error("failed to update product", "error", err, "product_id", id.String())
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("product updated", "product_id", id.String())

	return product, nil
}

func (p *ProductService) DeactivateProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := p.repo.GetById(ctx, id)
	if err != nil {
		p.log.Error("failed to deactivate product", "error", err, "product_id", id.String())
		return nil, errors.Unwrap(err)
	}

	if err := p.repo.Deactivate(ctx, id); err != nil {
		p.log.Error("failed to deactivate product", "error", err, "product_id", id.String())
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("product deactivated", "product_id", id.String())

	return product, nil
}

func (p *ProductService) GetById(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := p.repo.GetById(ctx, id)
	if err != nil {
		p.log.Error("failed to get product", "error", err, "product_id", id.String())
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("product retrieved", "product_id", id.String())

	return product, nil
}

func (p *ProductService) SearchProducts(ctx context.Context, filters map[string]any) ([]*domain.Product, error) {
	params := domain.NewSearchParams(filters)

	if err := params.Validate(); err != nil {
		p.log.Error("failed to search products", "error", err)
		return nil, err
	}

	products, err := p.repo.Search(ctx, params)
	if err != nil {
		p.log.Error("failed to search products", "error", err)
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("products retrieved", "count", len(products))

	return products, nil
}
