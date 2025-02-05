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
		Columns("id", "name", "description", "category", "price", "created_at", "updated_at").
		Values(product.ID, product.Name, product.Desc, product.Category, product.Price, product.CreatedAt, product.UpdatedAt).
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

func (p ProductRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	const op = "repository.ProductRepository.Get"

	selectBuilder := sq.Select("id", "name", "description", "category", "price", "created_at", "updated_at").
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

func (p ProductRepository) GetAll(ctx context.Context) ([]*domain.Product, error) {
	const op = "repository.ProductRepository.GetAll"

	selectBuilder := sq.Select("id", "name", "description", "category", "price", "created_at", "updated_at").
		From(productsTableName).
		PlaceholderFormat(sq.Dollar)

	query, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var products []*domain.Product

	rows, err := p.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for rows.Next() {
		var product domain.Product

		if err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Desc,
			&product.Category,
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

func (p ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	const op = "repository.ProductRepository.Update"

	updateBuilder := sq.Update(productsTableName).
		Set("name", product.Name).
		Set("description", product.Desc).
		Set("category", product.Category).
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

func (p ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "repository.ProductRepository.Delete"

	deleteBuilder := sq.Delete(productsTableName).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := deleteBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = p.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
