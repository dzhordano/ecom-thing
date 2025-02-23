package pg

import (
	"context"
	"errors"
	"fmt"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	itemsTable = "items"
)

type PGRepostory struct {
	db *pgxpool.Pool
}

func NewPGRepository(ctx context.Context, db *pgxpool.Pool) *PGRepostory {
	return &PGRepostory{db: db}
}

func (r *PGRepostory) GetItem(ctx context.Context, id string) (*domain.Item, error) {
	query := fmt.Sprintf(
		`SELECT available_quantity, reserved_quantity FROM %s WHERE product_id = $1`,
		itemsTable)

	var item domain.Item
	if err := r.db.QueryRow(ctx, query, id).Scan(&item.AvailableQuantity, &item.ReservedQuantity); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (r *PGRepostory) SetItem(ctx context.Context, id string, availableQuantity, reservedQuantity uint64) error {
	// Insert. If exists -> update.
	query := fmt.Sprintf(
		`INSERT INTO %s (product_id, available_quantity, reserved_quantity) VALUES ($1, $2, $3)
		ON CONFLICT (product_id) DO UPDATE SET available_quantity = $2, reserved_quantity = $3`,
		itemsTable)

	_, err := r.db.Exec(ctx, query, id, availableQuantity, reservedQuantity)
	if err != nil {
		return err
	}

	return nil
}
