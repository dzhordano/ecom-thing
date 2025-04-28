package inventory

import (
	"context"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"log"
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	api "github.com/dzhordano/ecom-thing/services/order/pkg/third_party/inventory/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type ClientOption func(*inventoryClient)

func WithTracing(tp *tracesdk.TracerProvider) ClientOption {
	return func(s *inventoryClient) {
		s.tp = tp
	}
}

type inventoryClient struct {
	c    api.InventoryServiceClient
	addr string
	tp   *tracesdk.TracerProvider
}

// NewInventoryClient creates new inventory client.
//
// Dials to the given address.
func NewInventoryClient(addr string, opts ...ClientOption) interfaces.InventoryService {
	client := inventoryClient{}

	for _, o := range opts {
		o(&client)
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

	if client.tp != nil {
		sOpts = append(sOpts,
			grpc.WithStatsHandler(
				otelgrpc.NewClientHandler(
					otelgrpc.WithPropagators(propagation.TraceContext{}),
					otelgrpc.WithTracerProvider(client.tp),
				),
			),
		)
	}

	// FIXME апогей хардкода..... или норм? (:(:(:(:
	//
	// Еще, для идеала надо бы retry-логику намутить еще
	conn, err := grpc.NewClient(
		addr,
		sOpts...,
	)
	if err != nil {
		log.Printf("failed to create grpc_server client: %v", err)
		return nil
	}

	client.c = api.NewInventoryServiceClient(conn)
	client.addr = addr

	return &client
}

// IsReservable implements interfaces.InventoryService.
func (i *inventoryClient) IsReservable(ctx context.Context, items map[string]uint64) (bool, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("order").Start(ctx, "IsReservable")
	defer span.End()

	span.AddEvent(
		"parse items",
		trace.WithAttributes(
			attribute.Int("items count", len(items)),
		),
	)

	protoItems := make([]*api.ItemOP, 0, len(items))

	for id := range items {
		protoItems = append(protoItems, &api.ItemOP{
			ProductId: id,
			Quantity:  items[id],
		})
	}

	span.AddEvent("performing rpc")

	resp, err := i.c.IsReservable(ctx, &api.IsReservableRequest{
		Items: protoItems,
	})
	if err != nil {
		return false, err
	}

	span.AddEvent("got response")

	return resp.IsReservable, nil
}
