package inventory

import (
	"context"
	"log"
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	inventory_v1 "github.com/dzhordano/ecom-thing/services/order/pkg/api/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// TODO ТУТА ХОДИМ В ИНВЕТАРЬ ЗА РЕЗЕРВАЦИЕЙ ПРЯМООО
type inventoryClient struct {
	c    inventory_v1.InventoryServiceClient
	addr string
}

func NewInventoryClient(addr string) interfaces.InventoryService {
	// FIXME апогей хардкода..... или норм? (:(:(:(:
	//
	// Еще, для идеала надо бы retry-логику намутить еще
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithIdleTimeout(5*time.Second),
		grpc.WithConnectParams(
			grpc.ConnectParams{
				Backoff:           backoff.DefaultConfig,
				MinConnectTimeout: 5 * time.Second,
			},
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    30 * time.Second,
			Timeout: 10 * time.Second,
		}),
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

// IsReservable implements interfaces.InventoryService.
func (i *inventoryClient) IsReservable(ctx context.Context, items map[string]uint64) (bool, error) {
	protoItems := make([]*inventory_v1.ItemOP, 0, len(items))

	for id := range items {
		protoItems = append(protoItems, &inventory_v1.ItemOP{
			ProductId: id,
			Quantity:  items[id],
		})
	}

	resp, err := i.c.IsReservable(ctx, &inventory_v1.IsReservableRequest{
		Items: protoItems,
	})
	if err != nil {
		return false, err
	}

	return resp.IsReservable, nil
}
