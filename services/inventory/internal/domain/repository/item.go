package repository

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
)

type ItemRepository interface {
	GetItem(ctx context.Context, id string) (*domain.Item, error)
	SetItem(ctx context.Context, id string, availableQuantity, reservedQuantity uint64) error
}
