package grpc_server

import (
	"context"
	"fmt"
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"github.com/dzhordano/ecom-thing/services/order/internal/interfaces/grpc_server/converter"
	api "github.com/dzhordano/ecom-thing/services/order/pkg/api/order/v1"
	"github.com/google/uuid"
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
	items, err := converter.RPCItemsToDomain(req.GetItems())
	if err != nil {
		return nil, err
	}

	fmt.Println("DESCRIPTION", req.GetDescription())

	info := dto.CreateOrderRequest{
		Description:     req.GetDescription(),
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
	items, err := converter.RPCItemsToDomain(req.GetItems())
	if err != nil {
		return nil, err
	}

	oid, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, domain.ErrInvalidUUID
	}

	var t time.Time
	if req.DeliveryDate.IsValid() {
		t = req.DeliveryDate.AsTime()
	}

	info := dto.UpdateOrderRequest{
		OrderID:         oid,
		Description:     req.Description,
		Status:          req.Status,
		TotalPrice:      req.TotalPrice,
		PaymentMethod:   req.PaymentMethod,
		DeliveryMethod:  req.DeliveryMethod,
		DeliveryAddress: req.DeliveryAddress,
		DeliveryDate:    t,
		Items:           items,
	}

	o, err := h.service.UpdateOrder(ctx, info)
	if err != nil {
		return nil, err
	}

	resp := &api.UpdateOrderResponse{
		Order: converter.FromDomainToProto_Order(o),
	}

	return resp, nil
}

func (h *ItemHandler) DeleteOrder(ctx context.Context, req *api.DeleteOrderRequest) (*api.DeleteOrderResponse, error) {
	oid, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, domain.ErrInvalidUUID
	}

	err = h.service.DeleteOrder(ctx, oid)
	if err != nil {
		return nil, err
	}

	return &api.DeleteOrderResponse{}, nil
}

func (h *ItemHandler) SearchOrders(ctx context.Context, req *api.SearchOrdersRequest) (*api.SearchOrdersResponse, error) {
	orders, err := h.service.SearchOrders(ctx, map[string]any{
		"limit":            req.Limit,
		"offset":           req.Offset,
		"query":            req.Query,
		"description":      req.Description,
		"status":           req.Status,
		"currency":         req.Currency,
		"minPrice":         req.MinPrice,
		"maxPrice":         req.MaxPrice,
		"deliveryMethod":   req.DeliveryMethod,
		"paymentMethod":    req.PaymentMethod,
		"deliveryAddress":  req.DeliveryAddress,
		"deliveryDateFrom": timeFromProtoIfNotZero(req.DeliveryDateFrom),
		"deliveryDateTo":   timeFromProtoIfNotZero(req.DeliveryDateTo),
		"minItemsAmount":   req.MinItemsAmount,
		"maxItemsAmount":   req.MaxItemsAmount,
	})
	if err != nil {
		return nil, err
	}

	resp := &api.SearchOrdersResponse{
		Orders: converter.FromDomainToProto_Orders(orders),
	}

	return resp, nil
}

func (h *ItemHandler) CompleteOrder(ctx context.Context, req *api.CompleteOrderRequest) (*api.CompleteOrderResponse, error) {
	oid, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, domain.ErrInvalidUUID
	}

	err = h.service.CompleteOrder(ctx, oid)
	if err != nil {
		return nil, err
	}

	return &api.CompleteOrderResponse{}, nil
}

func (h *ItemHandler) CancelOrder(ctx context.Context, req *api.CancelOrderRequest) (*api.CancelOrderResponse, error) {
	oid, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, domain.ErrInvalidUUID
	}

	err = h.service.CancelOrder(ctx, oid)
	if err != nil {
		return nil, err
	}

	return &api.CancelOrderResponse{}, nil
}

// This func checks if input time of proto type is zero. If it - returns nil time.
//
// Using because when you use t.AsTime() it applies 1970-01-01 00:00:00 +0000 UTC. Idk why.
// TODO видимо отчет из-за устройства чисел..? узнать почему.
func timeFromProtoIfNotZero(t *timestamppb.Timestamp) time.Time {
	if t == nil || (t.Nanos == 0 && t.Seconds == 0) {
		return time.Time{}
	}
	return t.AsTime()
}
