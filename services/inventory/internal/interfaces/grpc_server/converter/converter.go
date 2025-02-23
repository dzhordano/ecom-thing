package converter

import (
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	api "github.com/dzhordano/ecom-thing/services/inventory/pkg/api/inventory/v1"
)

func ItemToProto(item *domain.Item) *api.Item {
	return &api.Item{
		ProductId:         item.ProductID.String(),
		AvailableQuantity: item.AvailableQuantity,
		ReservedQuantity:  item.ReservedQuantity,
	}
}
