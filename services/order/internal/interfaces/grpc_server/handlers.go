package grpc_server

import (
	api "github.com/dzhordano/ecom-thing/services/order/pkg/api/order/v1"
)

type ItemHandler struct {
	api.UnimplementedOrderServiceServer
}
