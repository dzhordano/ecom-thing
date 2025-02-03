package pg

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) repository.ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (p ProductRepository) Save(ctx context.Context, product *domain.Product) error {
	// TODO implement me
	panic("implement me")
}

func (p ProductRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	// TODO implement me
	panic("implement me")
}

func (p ProductRepository) GetAll(ctx context.Context) ([]*domain.Product, error) {
	// TODO implement me
	panic("implement me")
}

func (p ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	// TODO implement me
	panic("implement me")
}

func (p ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// TODO implement me
	panic("implement me")
}
