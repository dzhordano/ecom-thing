package grpc_server

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	api "github.com/dzhordano/ecom-thing/services/order/pkg/api/order/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ItemHandler struct {
	api.UnimplementedOrderServiceServer
	service interfaces.OrderService
}

func NewItemHandler(s interfaces.OrderService) *ItemHandler {
	return &ItemHandler{
		service: s,
	}
}

func (h *ItemHandler) CreateOrder(ctx context.Context, req *api.CreateOrderRequest) (*api.CreateOrderResponse, error) {
	items, err := dto.RPCItemsToDomain(req.GetItems())
	if err != nil {
		return nil, err
	}

	info := dto.CreateOrderRequest{
		Currency:        req.GetCurrency(),
		TotalPrice:      req.GetTotalPrice(),
		Coupon:          req.GetCoupon(),
		PaymentMethod:   req.GetPaymentMethod(),
		DeliveryMethod:  req.GetDeliveryMethod(),
		DeliveryAddress: req.GetDeliveryAddress(),
		DeliveryDate:    req.GetDeliveryDate().AsTime(),
		Items:           items,
	}

	order, err := h.service.CreateOrder(ctx, info)
	if err != nil {
		return nil, err
	}

	resp := &api.CreateOrderResponse{
		Order: &api.Order{
			OrderId:         order.ID.String(),
			UserId:          order.UserID.String(),
			Status:          order.Status.String(),
			Currency:        order.Currency.String(),
			TotalPrice:      order.TotalPrice,
			Coupon:          req.GetCoupon(),
			PaymentMethod:   order.PaymentMethod.String(),
			DeliveryMethod:  order.DeliveryMethod.String(),
			DeliveryAddress: order.DeliveryAddress,
			DeliveryDate: &timestamppb.Timestamp{
				Seconds: order.DeliveryDate.Unix(),
				Nanos:   int32(order.DeliveryDate.Nanosecond()),
			},
			Items: req.Items,
			CreatedAt: &timestamppb.Timestamp{
				Seconds: order.CreatedAt.Unix(),
				Nanos:   int32(order.CreatedAt.Nanosecond()),
			},
			UpdatedAt: &timestamppb.Timestamp{
				Seconds: order.UpdatedAt.Unix(),
				Nanos:   int32(order.UpdatedAt.Nanosecond()),
			},
		},
	}

	return resp, nil
}

func (h *ItemHandler) GetOrder(ctx context.Context, req *api.GetOrderRequest) (*api.GetOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (h *ItemHandler) ListOrders(ctx context.Context, req *api.ListOrdersRequest) (*api.ListOrdersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (h *ItemHandler) UpdateOrder(ctx context.Context, req *api.UpdateOrderRequest) (*api.UpdateOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (h *ItemHandler) DeleteOrder(ctx context.Context, req *api.DeleteOrderRequest) (*api.DeleteOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (h *ItemHandler) SearchOrders(ctx context.Context, req *api.SearchOrdersRequest) (*api.SearchOrdersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (h *ItemHandler) CompleteOrder(ctx context.Context, req *api.CompleteOrderRequest) (*api.CompleteOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (h *ItemHandler) CancelOrder(ctx context.Context, req *api.CancelOrderRequest) (*api.CancelOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
