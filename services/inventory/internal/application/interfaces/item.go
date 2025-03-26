package interfaces

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"github.com/google/uuid"
)

type ItemService interface {
	GetItem(ctx context.Context, id uuid.UUID) (*domain.Item, error)
	IsReservable(ctx context.Context, items map[string]uint64) (bool, error)
	SetItemWithOp(ctx context.Context, id uuid.UUID, quantity uint64, op string) error
	SetItemsWithOp(ctx context.Context, items map[string]uint64, op string) error
}
