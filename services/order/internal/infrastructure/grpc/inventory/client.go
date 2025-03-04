package inventory

import (
	"context"
	"log"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	inventory_v1 "github.com/dzhordano/ecom-thing/services/order/pkg/api/inventory/v1"
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

// SetItemsWithOp implements interfaces.InventoryService.
func (i *inventoryClient) SetItemsWithOp(ctx context.Context, items map[string]uint64, op string) error {
	_, err := i.c.SetItems(ctx, &inventory_v1.SetItemsRequest{
		OperationType: operationToProtoType(op),
		Items:         convertMapToItemOPs(items),
	})
	return err
}

// Convert map with UUID's and quantities to []*ItemOP.
//
// This does NOT check if the UUID is valid.
func convertMapToItemOPs(items map[string]uint64) []*inventory_v1.ItemOP {
	ops := make([]*inventory_v1.ItemOP, 0, len(items))
	for id, quantity := range items {
		ops = append(ops, &inventory_v1.ItemOP{
			ProductId: id,
			Quantity:  quantity,
		})
	}
	return ops
}

func operationToProtoType(op string) inventory_v1.OperationType {
	switch op {
	case "add":
		return inventory_v1.OperationType_OPERATION_TYPE_ADD
	case "sub":
		return inventory_v1.OperationType_OPERATION_TYPE_SUB
	case "lock":
		return inventory_v1.OperationType_OPERATION_TYPE_LOCK
	case "unlock":
		return inventory_v1.OperationType_OPERATION_TYPE_UNLOCK
	case "sub_locked":
		return inventory_v1.OperationType_OPERATION_TYPE_SUB_LOCKED
	default:
		return inventory_v1.OperationType_OPERATION_TYPE_ADD
	}
}
