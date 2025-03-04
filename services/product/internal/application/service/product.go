package service

import (
	"context"
	"errors"

	"github.com/dzhordano/ecom-thing/services/product/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain/repository"
	"github.com/dzhordano/ecom-thing/services/product/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductService struct {
	log  logger.BaseLogger
	repo repository.ProductRepository
}

func NewProductService(log logger.BaseLogger, repo repository.ProductRepository) interfaces.ProductService {
	return &ProductService{
		log:  log,
		repo: repo,
	}
}

func (p *ProductService) CreateProduct(ctx context.Context, name, description, category string, price float64) (*domain.Product, error) {
	userId, err := uuid.NewUUID()
	if err != nil {
		p.log.Error("failed to create grpc", zap.Error(err))
		return nil, err
	}

	product, err := domain.NewValidatedProduct(userId, name, description, category, price)
	if err != nil {
		p.log.Error("failed to create grpc", zap.Error(err))
		return nil, err
	}

	if err := p.repo.Save(ctx, product); err != nil {
		p.log.Error("failed to save grpc", zap.Error(err))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("grpc created", zap.String("id", product.ID.String()))

	return product, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, name, description, category string, isActive bool, price float64) (*domain.Product, error) {
	product, err := p.repo.GetById(ctx, id)
	if err != nil {
		p.log.Error("failed to update grpc", zap.Error(err))
		return nil, errors.Unwrap(err)
	}

	product.Update(name, description, category, isActive, price)

	if err := product.Validate(); err != nil {
		p.log.Error("failed to update grpc", zap.Error(err))
		return nil, err
	}

	if err := p.repo.Update(ctx, product); err != nil {
		p.log.Error("failed to update grpc", zap.Error(err))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("grpc updated", zap.String("id", product.ID.String()))

	return product, nil
}

func (p *ProductService) DeactivateProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := p.repo.GetById(ctx, id)
	if err != nil {
		p.log.Error("failed to deactivate grpc", zap.Error(err))
		return nil, errors.Unwrap(err)
	}

	if err := p.repo.Deactivate(ctx, id); err != nil {
		p.log.Error("failed to deactivate grpc", zap.Error(err))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("grpc deactivated", zap.String("id", product.ID.String()))

	return product, nil
}

func (p *ProductService) GetById(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := p.repo.GetById(ctx, id)
	if err != nil {
		p.log.Error("failed to get grpc", zap.Error(err))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("grpc retrieved", zap.Error(err))

	return product, nil
}

func (p *ProductService) SearchProducts(ctx context.Context, filters map[string]any) ([]*domain.Product, error) {
	params := domain.NewSearchParams(filters)

	if err := params.Validate(); err != nil {
		p.log.Error("failed to search products", zap.Error(err))
		return nil, err
	}

	products, err := p.repo.Search(ctx, params)
	if err != nil {
		p.log.Error("failed to search products", zap.Error(err))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("products retrieved", zap.Int("count", len(products)))

	return products, nil
}
