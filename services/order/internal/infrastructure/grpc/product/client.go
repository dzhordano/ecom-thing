package product

import (
	"context"
	"log"
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	product_v1 "github.com/dzhordano/ecom-thing/services/order/pkg/api/product/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// TODO ТУТА ХОДИМ К ПРОДУКТАМ ПРЯМО

type productClient struct {
	c    product_v1.ProductServiceClient
	addr string
}

func NewProductClient(addr string) interfaces.ProductService {
	// FIXME апять многа цифар "рандомных"
	//
	// также ретрай для идеала нужон
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

	return &productClient{
		c:    product_v1.NewProductServiceClient(conn),
		addr: addr,
	}
}

func (c *productClient) GetProductInfo(ctx context.Context, orderId uuid.UUID) (float64, bool, error) {
	resp, err := c.c.GetProduct(ctx, &product_v1.GetProductRequest{
		Id: orderId.String(),
	})
	if err != nil {
		return 0, false, err
	}

	return resp.Product.Price, resp.Product.IsActive, nil
}
