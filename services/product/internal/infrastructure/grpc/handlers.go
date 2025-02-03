package grpc

import (
	"github.com/dzhordano/ecom-thing/services/product/internal/application/interfaces"
	productv1 "github.com/dzhordano/ecom-thing/services/product/pkg/grpc/product/v1"
)

type ProductHandler struct {
	productv1.UnimplementedProductServiceV1Server
	service interfaces.ProductService
}

func NewProductHandler(service interfaces.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}
