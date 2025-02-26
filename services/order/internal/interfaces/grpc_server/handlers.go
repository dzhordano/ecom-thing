package grpc_server

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"github.com/dzhordano/ecom-thing/services/order/internal/interfaces/grpc_server/converter"
	api "github.com/dzhordano/ecom-thing/services/order/pkg/api/order/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		Order: converter.FromDomainToProto_OrderWItems(order, req.Items),
	}

	return resp, nil
}

func (h *ItemHandler) GetOrder(ctx context.Context, req *api.GetOrderRequest) (*api.GetOrderResponse, error) {
	orderId, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, domain.ErrInvalidUUID
	}

	order, err := h.service.GetById(ctx, orderId)
	if err != nil {
		return nil, err
	}

	resp := &api.GetOrderResponse{Order: converter.FromDomainToProto_Order(order)}

	return resp, nil
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
