package grpc_server

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

type OrderHandler struct {
	api.UnimplementedOrderServiceServer
	service interfaces.OrderService
}

func NewOrderHandler(s interfaces.OrderService) *OrderHandler {
	return &OrderHandler{
		service: s,
	}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *api.CreateOrderRequest) (*api.CreateOrderResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse items",
		trace.WithAttributes(
			attribute.Int("count", len(req.GetItems())),
		),
	)

	items, err := converter.RPCItemsToDomain(req.GetItems())
	if err != nil {
		return nil, err
	}

	info := dto.CreateOrderRequest{
		Description:     req.GetDescription(),
		Currency:        req.GetCurrency(),
		Coupon:          req.GetCoupon(),
		PaymentMethod:   req.GetPaymentMethod(),
		DeliveryMethod:  req.GetDeliveryMethod(),
		DeliveryAddress: req.GetDeliveryAddress(),
		DeliveryDate:    req.GetDeliveryDate().AsTime(),
		Items:           items,
	}

	span.AddEvent("call service")

	order, err := h.service.CreateOrder(ctx, info)
	if err != nil {
		return nil, err
	}

	span.AddEvent("order created",
		trace.WithAttributes(
			attribute.Stringer("order_id", order.ID),
		),
	)

	resp := &api.CreateOrderResponse{
		Order: converter.FromDomainToProto_OrderWItems(order, req.Items),
	}

	return resp, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *api.GetOrderRequest) (*api.GetOrderResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("order_id", req.GetOrderId()),
		),
	)

	orderId, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, domain.ErrInvalidUUID
	}

	span.AddEvent("call service")

	order, err := h.service.GetById(ctx, orderId)
	if err != nil {
		return nil, err
	}

	span.AddEvent("order retrieved",
		trace.WithAttributes(
			attribute.Stringer("order_id", order.ID),
		),
	)

	resp := &api.GetOrderResponse{Order: converter.FromDomainToProto_Order(order)}

	return resp, nil
}

// TODO: implement
func (h *OrderHandler) ListOrders(ctx context.Context, req *api.ListOrdersRequest) (*api.ListOrdersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (h *OrderHandler) UpdateOrder(ctx context.Context, req *api.UpdateOrderRequest) (*api.UpdateOrderResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse items",
		trace.WithAttributes(
			attribute.Int("count", len(req.GetItems())),
		),
	)

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

	span.AddEvent("call service")

	o, err := h.service.UpdateOrder(ctx, info)
	if err != nil {
		return nil, err
	}

	span.AddEvent("order updated",
		trace.WithAttributes(
			attribute.Stringer("order_id", o.ID),
		),
	)

	resp := &api.UpdateOrderResponse{
		Order: converter.FromDomainToProto_Order(o),
	}

	return resp, nil
}

func (h *OrderHandler) DeleteOrder(ctx context.Context, req *api.DeleteOrderRequest) (*api.DeleteOrderResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("order_id", req.GetOrderId()),
		),
	)

	oid, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, domain.ErrInvalidUUID
	}

	span.AddEvent("call service")

	err = h.service.DeleteOrder(ctx, oid)
	if err != nil {
		return nil, err
	}

	span.AddEvent("order deleted",
		trace.WithAttributes(
			attribute.Stringer("order_id", oid),
		),
	)

	return &api.DeleteOrderResponse{}, nil
}

func (h *OrderHandler) SearchOrders(ctx context.Context, req *api.SearchOrdersRequest) (*api.SearchOrdersResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("call service",
		trace.WithAttributes(
			attribute.Stringer("req", req),
		),
	)

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

	span.AddEvent("orders found",
		trace.WithAttributes(
			attribute.Int("count", len(orders)),
		),
	)

	resp := &api.SearchOrdersResponse{
		Orders: converter.FromDomainToProto_Orders(orders),
	}

	return resp, nil
}

func (h *OrderHandler) CompleteOrder(ctx context.Context, req *api.CompleteOrderRequest) (*api.CompleteOrderResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("order_id", req.GetOrderId()),
		),
	)

	oid, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, domain.ErrInvalidUUID
	}

	span.AddEvent("call service")

	err = h.service.CompleteOrder(ctx, oid)
	if err != nil {
		return nil, err
	}

	span.AddEvent("order completed")

	return &api.CompleteOrderResponse{}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *api.CancelOrderRequest) (*api.CancelOrderResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("order_id", req.GetOrderId()),
		),
	)

	oid, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, domain.ErrInvalidUUID
	}

	span.AddEvent("call service")

	err = h.service.CancelOrder(ctx, oid)
	if err != nil {
		return nil, err
	}

	span.AddEvent("order canceled")

	return &api.CancelOrderResponse{}, nil
}

// This func checks if input time of proto type is zero. If so - returns nil time.
//
// Because when you use t.AsTime() it applies 1970-01-01 00:00:00 +0000 UTC as zero value due to protobuf implementation.
func timeFromProtoIfNotZero(t *timestamppb.Timestamp) time.Time {
	if t == nil || (t.Nanos == 0 && t.Seconds == 0) {
		return time.Time{}
	}
	return t.AsTime()
}
