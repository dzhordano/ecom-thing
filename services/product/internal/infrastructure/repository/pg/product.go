package pg

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	productsTableName = "products"
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
	const op = "repository.ProductRepository.Save"

	insertBuilder := sq.Insert(productsTableName).
		Columns("id", "name", "description", "category", "is_active", "price", "created_at", "updated_at").
		Values(product.ID, product.Name, product.Desc, product.Category, product.IsActive, product.Price, product.CreatedAt, product.UpdatedAt).
		PlaceholderFormat(sq.Dollar)

	query, args, err := insertBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = p.db.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return fmt.Errorf("%s: %w", op, domain.ErrProductAlreadyExists)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	const op = "repository.ProductRepository.Update"

	updateBuilder := sq.Update(productsTableName).
		Set("name", product.Name).
		Set("description", product.Desc).
		Set("category", product.Category).
		Set("is_active", product.IsActive).
		Set("price", product.Price).
		Set("updated_at", product.UpdatedAt).
		Where(sq.Eq{"id": product.ID}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = p.db.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return fmt.Errorf("%s: %w", op, domain.ErrProductAlreadyExists)
			}
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p ProductRepository) Deactivate(ctx context.Context, id uuid.UUID) error {
	const op = "repository.ProductRepository.Deactivate"

	updateBuilder := sq.Update(productsTableName).
		Set("is_active", false).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = p.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p ProductRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	const op = "repository.ProductRepository.GetById"

	selectBuilder := sq.Select("id", "name", "description", "category", "is_active", "price", "created_at", "updated_at").
		From(productsTableName).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var product domain.Product

	if err := p.db.QueryRow(ctx, query, args...).Scan(
		&product.ID,
		&product.Name,
		&product.Desc,
		&product.Category,
		&product.IsActive,
		&product.Price,
		&product.CreatedAt,
		&product.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrProductNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &product, nil
}

func (p ProductRepository) Search(ctx context.Context, params domain.SearchParams) ([]*domain.Product, error) {
	const op = "repository.ProductRepository.Search"

	selectBuilder := sq.Select("id", "name", "description", "category", "is_active", "price", "created_at", "updated_at").
		From(productsTableName).
		PlaceholderFormat(sq.Dollar).
		Limit(params.Limit).
		Offset(params.Offset)

	if params.Query != nil {
		selectBuilder = selectBuilder.Where(sq.Or{
			sq.Like{"name": fmt.Sprintf("%%%s%%", *params.Query)},
			sq.Like{"description": fmt.Sprintf("%%%s%%", *params.Query)},
			sq.Like{"category": fmt.Sprintf("%%%s%%", *params.Query)},
		})
	}

	if params.Category != nil {
		selectBuilder = selectBuilder.Where(sq.Eq{"category": *params.Category})
	}

	if params.MinPrice != nil {
		selectBuilder = selectBuilder.Where(sq.GtOrEq{"price": *params.MinPrice})
	}

	if params.MaxPrice != nil {
		selectBuilder = selectBuilder.Where(sq.LtOrEq{"price": *params.MaxPrice})
	}

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := p.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var products []*domain.Product

	for rows.Next() {
		var product domain.Product

		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Desc,
			&product.Category,
			&product.IsActive,
			&product.Price,
			&product.CreatedAt,
			&product.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		products = append(products, &product)
	}

	return products, nil
}
