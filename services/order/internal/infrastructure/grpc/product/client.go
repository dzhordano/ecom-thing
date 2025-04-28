package product

import (
	"context"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"log"
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	api "github.com/dzhordano/ecom-thing/services/order/pkg/third_party/product/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type ClientOption func(*productClient)

func WithTracing(tp *tracesdk.TracerProvider) ClientOption {
	return func(s *productClient) {
		s.tp = tp
	}
}

type productClient struct {
	c    api.ProductServiceClient
	addr string
	tp   *tracesdk.TracerProvider
}

func NewProductClient(addr string, opts ...ClientOption) interfaces.ProductService {
	s := &productClient{}

	for _, o := range opts {
		o(s)
	}

	sOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithIdleTimeout(5 * time.Second),
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
	}

	if s.tp != nil {
		sOpts = append(sOpts,
			grpc.WithStatsHandler(
				otelgrpc.NewClientHandler(
					otelgrpc.WithPropagators(propagation.TraceContext{}),
					otelgrpc.WithTracerProvider(s.tp),
				),
			),
		)
	}

	// FIXME апять многа цифар "рандомных"
	//
	// также ретрай для идеала нужон
	conn, err := grpc.NewClient(
		addr,
		sOpts...,
	)
	if err != nil {
		log.Printf("failed to create grpc_server client: %v", err)
		return nil
	}

	return &productClient{
		c:    api.NewProductServiceClient(conn),
		addr: addr,
	}
}

func (c *productClient) GetProductInfo(ctx context.Context, orderId uuid.UUID) (float64, bool, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("order").Start(ctx, "GetProductInfo")
	defer span.End()

	span.AddEvent("performing rpc")

	resp, err := c.c.GetProduct(ctx, &api.GetProductRequest{
		Id: orderId.String(),
	})
	if err != nil {
		return 0, false, err
	}

	span.AddEvent("got response")

	return resp.Product.Price, resp.Product.IsActive, nil
}
