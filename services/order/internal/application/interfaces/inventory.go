package interfaces

import "context"

type InventoryService interface {
	IsReservable(ctx context.Context, items map[string]uint64) (bool, error)
}
