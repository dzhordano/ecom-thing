package converter

import (
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	api "github.com/dzhordano/ecom-thing/services/product/pkg/api/product/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ProductToProto(product *domain.Product) *api.Product {
	return &api.Product{
		Id:       product.ID.String(),
		Name:     product.Name,
		Desc:     product.Desc,
		Category: product.Category,
		IsActive: product.IsActive,
		Price:    product.Price,
		//CreatedAt: timestamppb.New(product.CreatedAt),
		CreatedAt: &timestamppb.Timestamp{
			Seconds: product.CreatedAt.Unix(),
			Nanos:   int32(product.CreatedAt.Nanosecond()),
		},
		// UpdatedAt: timestamppb.New(product.UpdatedAt),
		UpdatedAt: &timestamppb.Timestamp{
			Seconds: product.UpdatedAt.Unix(),
			Nanos:   int32(product.UpdatedAt.Nanosecond()),
		},
	}
}

func ManyProductsToProto(products []*domain.Product) []*api.Product {
	var result []*api.Product

	for _, product := range products {
		result = append(result, ProductToProto(product))
	}

	return result
}
