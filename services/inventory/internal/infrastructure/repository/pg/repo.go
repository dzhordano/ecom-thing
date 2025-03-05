package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	itemsTable = "items"
)

type PGRepostory struct {
	db *pgxpool.Pool
}

func NewPGRepository(ctx context.Context, db *pgxpool.Pool) repository.ItemRepository {
	return &PGRepostory{db: db}
}

func (r *PGRepostory) GetItem(ctx context.Context, id string) (*domain.Item, error) {
	const op = "repository.PGRepostory.GetItem"

	query := fmt.Sprintf(
		`SELECT product_id, available_quantity, reserved_quantity FROM %s WHERE product_id = $1`,
		itemsTable)

	var item domain.Item
	if err := r.db.QueryRow(ctx, query, id).Scan(&item.ProductID, &item.AvailableQuantity, &item.ReservedQuantity); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrProductNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &item, nil
}

func (r *PGRepostory) SetItem(ctx context.Context, id string, availableQuantity, reservedQuantity uint64) error {
	const op = "repository.PGRepostory.SetItem"

	// Insert. If exists -> update.
	query := fmt.Sprintf(
		`INSERT INTO %s (product_id, available_quantity, reserved_quantity) VALUES ($1, $2, $3)
		ON CONFLICT (product_id) DO UPDATE SET available_quantity = $2, reserved_quantity = $3`,
		itemsTable)

	_, err := r.db.Exec(ctx, query, id, availableQuantity, reservedQuantity)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *PGRepostory) GetManyItems(ctx context.Context, ids []string) ([]*domain.Item, error) {
	const op = "repository.PGRepostory.GetManyItems"

	query := fmt.Sprintf(
		`SELECT product_id, available_quantity, reserved_quantity FROM %s WHERE product_id = ANY($1)`,
		itemsTable)

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var items []*domain.Item
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.ProductID, &item.AvailableQuantity, &item.ReservedQuantity); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		items = append(items, &item)
	}

	return items, nil
}

func (r *PGRepostory) SetManyItems(ctx context.Context, items []domain.Item) error {
	const op = "repository.PGRepostory.SetManyItems"

	return r.withTx(ctx, func(ctx context.Context, tx pgx.Tx) error {

		query := fmt.Sprintf(
			`UPDATE %s SET available_quantity = $2, reserved_quantity = $3 WHERE product_id = $1`,
			itemsTable)

		for _, item := range items {
			if _, err := tx.Exec(ctx, query, item.ProductID.String(), item.AvailableQuantity, item.ReservedQuantity); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		return nil
	})
}

func (r *PGRepostory) withTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(ctx, tx); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return err
		}
		return err
	}

	return tx.Commit(ctx)
}
