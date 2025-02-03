package grpc

import (
	productv1 "github.com/dzhordano/ecom-thing/services/product/pkg/grpc/product/v1"
)

type ProductHandler struct {
	productv1.UnimplementedProductServiceV1Server
}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}
