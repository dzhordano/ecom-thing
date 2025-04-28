package grpc_server

import (
	"context"
	"fmt"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/interfaces/grpc_server/converter"
	api "github.com/dzhordano/ecom-thing/services/inventory/pkg/api/inventory/v1"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ItemHandler struct {
	api.UnimplementedInventoryServiceServer
	service interfaces.ItemService
}

func NewItemHandler(service interfaces.ItemService) *ItemHandler {
	return &ItemHandler{
		service: service,
	}
}

func (h *ItemHandler) GetItem(ctx context.Context, req *api.GetItemRequest) (*api.GetItemResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("product_id", req.GetProductId()),
		),
	)

	itemId, err := parseUUID(req.GetProductId())
	if err != nil {
		return nil, err
	}

	span.AddEvent("call service")

	item, err := h.service.GetItem(ctx, itemId)
	if err != nil {
		return nil, err
	}

	protoItem := converter.ItemToProto(item)

	span.AddEvent(
		"got item",
		trace.WithAttributes(
			attribute.Stringer("item", protoItem),
		),
	)

	return &api.GetItemResponse{
		Item: protoItem,
	}, nil
}

func (h *ItemHandler) SetItem(ctx context.Context, req *api.SetItemRequest) (*api.SetItemResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent(
		"parse item",
		trace.WithAttributes(
			attribute.String("product_id", req.Item.GetProductId()),
			attribute.Stringer("op", req.OperationType),
		),
	)

	if req.Item.GetQuantity() == 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be greater than 0")
	}

	itemId, err := parseUUID(req.Item.GetProductId())
	if err != nil {
		return nil, err
	}

	span.AddEvent("call service")

	err = h.service.SetItemWithOp(ctx, itemId, req.Item.GetQuantity(), protoOpToString(req.OperationType))
	if err != nil {
		return nil, err
	}

	span.AddEvent("item set")

	return &api.SetItemResponse{}, nil
}

func (h *ItemHandler) SetItems(ctx context.Context, req *api.SetItemsRequest) (_ *api.SetItemsResponse, err error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent(
		"parse items",
		trace.WithAttributes(
			attribute.Int("items count", len(req.Items)),
			attribute.Stringer("op", req.OperationType),
		),
	)

	pItems := map[string]uint64{}
	for _, item := range req.Items {
		if err := validUUID(item.GetProductId()); err != nil {
			return nil, err
		}

		iq := item.GetQuantity()

		if iq == 0 {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("quantity for item %s must be greater than 0", item.GetProductId()))
		}

		pItems[item.ProductId] = item.GetQuantity()
	}

	span.AddEvent("call service")

	if err := h.service.SetItemsWithOp(ctx, pItems, protoOpToString(req.OperationType)); err != nil {
		return nil, err
	}

	span.AddEvent("items set")

	return &api.SetItemsResponse{}, nil
}

func (h *ItemHandler) IsReservable(ctx context.Context, req *api.IsReservableRequest) (_ *api.IsReservableResponse, err error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent(
		"parse items",
		trace.WithAttributes(
			attribute.Int("items count", len(req.Items)),
		),
	)

	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "no items provided")
	}

	items := map[string]uint64{}
	for i := range req.Items {
		if err := validUUID(req.Items[i].GetProductId()); err != nil {
			return nil, err
		}

		items[req.Items[i].ProductId] = req.Items[i].GetQuantity()
	}

	span.AddEvent("call service")

	resp, err := h.service.IsReservable(ctx, items)
	if err != nil {
		return nil, err
	}

	span.AddEvent(
		"got response",
		trace.WithAttributes(
			attribute.Bool("is_reservable", resp),
		),
	)

	return &api.IsReservableResponse{
		IsReservable: resp,
	}, nil
}

func parseUUID(id string) (uuid.UUID, error) {
	out, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, status.Error(codes.InvalidArgument, "invalid uuid")
	}

	return out, nil
}

func validUUID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid uuid")
	}

	return nil
}

func protoOpToString(op api.OperationType) string {
	switch op {
	case api.OperationType_OPERATION_TYPE_ADD:
		return domain.OperationAdd
	case api.OperationType_OPERATION_TYPE_SUB:
		return domain.OperationSub
	case api.OperationType_OPERATION_TYPE_LOCK:
		return domain.OperationLock
	case api.OperationType_OPERATION_TYPE_UNLOCK:
		return domain.OperationUnlock
	case api.OperationType_OPERATION_TYPE_SUB_LOCKED:
		return domain.OperationSubLocked
	default:
		return domain.OperationUnknown
	}
}
