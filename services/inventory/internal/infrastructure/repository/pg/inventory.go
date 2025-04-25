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

type InventoryRepository struct {
	db *pgxpool.Pool
}

func NewInventoryRepository(db *pgxpool.Pool) repository.ItemRepository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) GetItem(ctx context.Context, id string) (*domain.Item, error) {
	const op = "repository.InventoryRepository.GetItem"

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

func (r *InventoryRepository) SetItem(ctx context.Context, id string, availableQuantity, reservedQuantity uint64) error {
	const op = "repository.InventoryRepository.SetItem"

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

// TODO было бы идеально, если бы собиралась ошибка, которая укажет на ненайденные айди
func (r *InventoryRepository) GetManyItems(ctx context.Context, ids []string) ([]*domain.Item, error) {
	const op = "repository.InventoryRepository.GetManyItems"

	query := fmt.Sprintf(
		`SELECT product_id, available_quantity, reserved_quantity FROM %s WHERE product_id = ANY($1)`,
		itemsTable)

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

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

func (r *InventoryRepository) SetManyItems(ctx context.Context, items []domain.Item) error {
	const op = "repository.InventoryRepository.SetManyItems"

	return r.withTx(ctx, func(ctx context.Context, tx pgx.Tx) error {

		// TODO batch insert somehow?
		query := fmt.Sprintf(
			`INSERT INTO %s (product_id, available_quantity, reserved_quantity) VALUES ($1, $2, $3)
			ON CONFLICT (product_id) DO UPDATE SET available_quantity = $2, reserved_quantity = $3`,
			itemsTable)

		for _, item := range items {
			if _, err := tx.Exec(ctx, query, item.ProductID.String(), item.AvailableQuantity, item.ReservedQuantity); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		return nil
	})
}

func (r *InventoryRepository) withTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
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
