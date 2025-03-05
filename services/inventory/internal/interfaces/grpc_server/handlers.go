package grpc_server

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/interfaces/grpc_server/converter"
	api "github.com/dzhordano/ecom-thing/services/inventory/pkg/api/inventory/v1"
	"github.com/google/uuid"
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
	itemId, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	item, err := h.service.GetItem(ctx, itemId)
	if err != nil {
		return nil, err
	}

	return &api.GetItemResponse{
		Item: converter.ItemToProto(item),
	}, nil
}

func (h *ItemHandler) SetItem(ctx context.Context, req *api.SetItemRequest) (*api.SetItemResponse, error) {
	itemId, err := parseUUID(req.Item.GetProductId())
	if err != nil {
		return nil, err
	}

	err = h.service.SetItemWithOp(ctx, itemId, req.Item.GetQuantity(), protoOpToString(req.OperationType))
	if err != nil {
		return nil, err
	}

	return &api.SetItemResponse{}, nil
}

func (h *ItemHandler) SetItems(ctx context.Context, req *api.SetItemsRequest) (*api.SetItemsResponse, error) {
	pItems := map[string]uint64{}
	for _, item := range req.Items {
		if err := validUUID(item.GetProductId()); err != nil {
			return nil, err
		}

		pItems[item.ProductId] = item.GetQuantity()
	}

	err := h.service.SetItemsWithOp(ctx, pItems, protoOpToString(req.OperationType))
	if err != nil {
		return nil, err
	}

	return &api.SetItemsResponse{}, nil
}

func (h *ItemHandler) IsReservable(ctx context.Context, req *api.IsReservableRequest) (*api.IsReservableResponse, error) {
	items := map[string]uint64{}

	for i := range req.Items {
		if err := validUUID(req.Items[i].GetProductId()); err != nil {
			return nil, err
		}

		items[req.Items[i].ProductId] = req.Items[i].GetQuantity()
	}

	resp, err := h.service.IsReservable(ctx, items)
	if err != nil {
		return nil, err
	}

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
