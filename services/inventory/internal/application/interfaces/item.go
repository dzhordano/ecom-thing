package interfaces

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"github.com/google/uuid"
)

type ItemService interface {
	GetItem(ctx context.Context, id uuid.UUID) (*domain.Item, error)

	AddItemQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error
	SubItemQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error

	LockItemQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error
	UnlockItemQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error
	SubLockedItemQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error
}
