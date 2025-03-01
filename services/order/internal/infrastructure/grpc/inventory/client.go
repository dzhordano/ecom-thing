package inventory

import (
	"context"
	"log"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	inventory_v1 "github.com/dzhordano/ecom-thing/services/order/pkg/api/inventory/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO ТУТА ХОДИМ В ИНВЕТАРЬ ЗА РЕЗЕРВАЦИЕЙ ПРЯМООО
type inventoryClient struct {
	c    inventory_v1.InventoryServiceClient
	addr string
}

func NewInventoryClient(addr string) interfaces.InventoryService {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("failed to create grpc client: %v", err)
		return nil
	}

	return &inventoryClient{
		c:    inventory_v1.NewInventoryServiceClient(conn),
		addr: addr,
	}
}

// ReserveQuantity implements interfaces.InventoryService.
func (i *inventoryClient) ReserveQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error {
	_, err := i.c.LockQuantity(ctx, &inventory_v1.LockQuantityRequest{
		Id:       id.String(),
		Quantity: quantity,
	})

	return err
}

// ReleaseQuantity implements interfaces.InventoryService.
func (i *inventoryClient) ReleaseQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error {
	_, err := i.c.UnlockQuantity(ctx, &inventory_v1.UnlockQuantityRequest{
		Id:       id.String(),
		Quantity: quantity,
	})

	return err
}

// SubReservedQuantity implements interfaces.InventoryService.
func (i *inventoryClient) SubReservedQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error {
	_, err := i.c.SubLockedQuantity(ctx, &inventory_v1.SubQuantityRequest{
		Id:       id.String(),
		Quantity: quantity,
	})

	return err
}
