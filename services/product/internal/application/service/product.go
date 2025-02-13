package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/dzhordano/ecom-thing/services/product/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain/repository"
	"github.com/google/uuid"
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
		p.log.Error("failed to create grpc", slog.String("error", err.Error()))
		return nil, err
	}

	product, err := domain.NewValidatedProduct(userId, name, description, category, price)
	if err != nil {
		p.log.Error("failed to create grpc", slog.String("error", err.Error()))
		return nil, err
	}

	if err := p.repo.Save(ctx, product); err != nil {
		p.log.Error("failed to save grpc", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("grpc created", "id", product.ID)

	return product, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, id uuid.UUID, name, description, category string, isActive bool, price float64) (*domain.Product, error) {
	product, err := p.repo.GetById(ctx, id)
	if err != nil {
		p.log.Error("failed to update grpc", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	product.Update(name, description, category, isActive, price)

	if err := product.Validate(); err != nil {
		p.log.Error("failed to update grpc", slog.String("error", err.Error()))
		return nil, err
	}

	if err := p.repo.Update(ctx, product); err != nil {
		p.log.Error("failed to update grpc", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("grpc updated", "id", product.ID)

	return product, nil
}

func (p *ProductService) DeactivateProduct(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := p.repo.GetById(ctx, id)
	if err != nil {
		p.log.Error("failed to deactivate grpc", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	if err := p.repo.Deactivate(ctx, id); err != nil {
		p.log.Error("failed to deactivate grpc", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("grpc deactivated", "id", product.ID)

	return product, nil
}

func (p *ProductService) GetById(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	product, err := p.repo.GetById(ctx, id)
	if err != nil {
		p.log.Error("failed to get grpc", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("grpc retrieved", "id", product.ID)

	return product, nil
}

func (p *ProductService) SearchProducts(ctx context.Context, filters map[string]any, limit, offset uint64) ([]*domain.Product, error) {
	q, ok := filters["query"].(*string)
	if !ok {
		q = nil
	}

	c, ok := filters["category"].(*string)
	if !ok {
		c = nil
	}

	mn, ok := filters["minPrice"].(*float64)
	if !ok {
		mn = nil
	}

	mx, ok := filters["maxPrice"].(*float64)
	if !ok {
		mx = nil
	}

	options := domain.NewSearchOptions(q, c, mn, mx)

	if err := options.Validate(); err != nil {
		p.log.Error("failed to search products", slog.String("error", err.Error()))
		return nil, err
	}

	products, err := p.repo.Search(ctx, options, limit, offset)
	if err != nil {
		p.log.Error("failed to search products", slog.String("error", err.Error()))
		return nil, errors.Unwrap(err)
	}

	p.log.Debug("products retrieved", "count", len(products))

	return products, nil
}
