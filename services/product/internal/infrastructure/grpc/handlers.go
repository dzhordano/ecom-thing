package grpc

import (
	product_v1 "github.com/dzhordano/ecom-thing/services/product/pkg/grpc/product/v1"
)

type ProductHandler struct {
	product_v1.UnimplementedProductServiceV1Server
}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}
