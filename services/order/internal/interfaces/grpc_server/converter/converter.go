package converter

import (
	"fmt"

	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	order_v1 "github.com/dzhordano/ecom-thing/services/order/pkg/api/order/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func FromDomainToProto_Order(order *domain.Order) *order_v1.Order {
	fmt.Printf("ORDER %+v\n\n\n\n\n\n", order)

	return &order_v1.Order{
		OrderId:         order.ID.String(),
		UserId:          order.UserID.String(),
		Description:     order.Description,
		Status:          order.Status.String(),
		Currency:        order.Currency.String(),
		TotalPrice:      order.TotalPrice,
		PaymentMethod:   order.PaymentMethod.String(),
		DeliveryMethod:  order.DeliveryMethod.String(),
		DeliveryAddress: order.DeliveryAddress,
		DeliveryDate: &timestamppb.Timestamp{
			Seconds: order.DeliveryDate.Unix(),
			Nanos:   int32(order.DeliveryDate.Nanosecond()),
		},
		Items: FromDomainToProto_Items(order.Items),
		CreatedAt: &timestamppb.Timestamp{
			Seconds: order.CreatedAt.Unix(),
			Nanos:   int32(order.CreatedAt.Nanosecond()),
		},
		UpdatedAt: &timestamppb.Timestamp{
			Seconds: order.UpdatedAt.Unix(),
			Nanos:   int32(order.UpdatedAt.Nanosecond()),
		},
	}
}

func FromDomainToProto_Orders(orders []*domain.Order) []*order_v1.Order {
	var result []*order_v1.Order
	for _, order := range orders {
		result = append(result, FromDomainToProto_Order(order))
	}
	return result
}

// FromDomainToProto_OrderWItems is same as FromDomainToProto_Order but with items for better performance.
func FromDomainToProto_OrderWItems(order *domain.Order, items []*order_v1.Item) *order_v1.Order {
	return &order_v1.Order{
		OrderId:         order.ID.String(),
		UserId:          order.UserID.String(),
		Description:     order.Description,
		Status:          order.Status.String(),
		Currency:        order.Currency.String(),
		TotalPrice:      order.TotalPrice,
		PaymentMethod:   order.PaymentMethod.String(),
		DeliveryMethod:  order.DeliveryMethod.String(),
		DeliveryAddress: order.DeliveryAddress,
		DeliveryDate: &timestamppb.Timestamp{
			Seconds: order.DeliveryDate.Unix(),
			Nanos:   int32(order.DeliveryDate.Nanosecond()),
		},
		Items: items,
		CreatedAt: &timestamppb.Timestamp{
			Seconds: order.CreatedAt.Unix(),
			Nanos:   int32(order.CreatedAt.Nanosecond()),
		},
		UpdatedAt: &timestamppb.Timestamp{
			Seconds: order.UpdatedAt.Unix(),
			Nanos:   int32(order.UpdatedAt.Nanosecond()),
		},
	}
}

func FromDomainToProto_Items(items []domain.Item) []*order_v1.Item {
	var result []*order_v1.Item
	for _, item := range items {
		result = append(result, &order_v1.Item{
			ItemId:   item.ProductID.String(),
			Quantity: item.Quantity,
		})
	}
	return result
}

func RPCItemsToDomain(items []*order_v1.Item) ([]domain.Item, error) {
	var result []domain.Item
	for _, item := range items {
		id, err := uuid.Parse(item.ItemId)
		if err != nil {
			return nil, domain.ErrInvalidUUID // FIXME too obscure
		}

		result = append(result, domain.Item{
			ProductID: id,
			Quantity:  item.GetQuantity(),
		})
	}
	return result, nil
}
