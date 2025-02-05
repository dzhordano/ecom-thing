package converter

import (
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	api "github.com/dzhordano/ecom-thing/services/product/pkg/grpc/product/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ProductToProto(product *domain.Product) *api.Product {
	return &api.Product{
		Id:        product.ID.String(),
		Name:      product.Name,
		Desc:      product.Desc,
		Category:  product.Category,
		Price:     product.Price,
		CreatedAt: timestamppb.New(product.CreatedAt),
		UpdatedAt: timestamppb.New(product.UpdatedAt),
	}
}

func ManyProductsToProto(products []*domain.Product) []*api.Product {
	var result []*api.Product

	for _, product := range products {
		result = append(result, ProductToProto(product))
	}

	return result
}
