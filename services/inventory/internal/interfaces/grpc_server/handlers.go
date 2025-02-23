package grpc_server

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/interfaces/grpc_server/converter"
	api "github.com/dzhordano/ecom-thing/services/inventory/pkg/api/inventory/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (h *ItemHandler) AddQuantity(ctx context.Context, req *api.AddQuantityRequest) (*emptypb.Empty, error) {
	itemId, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	if err = h.service.AddItemQuantity(ctx, itemId, req.GetQuantity()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ItemHandler) SubQuantity(ctx context.Context, req *api.SubQuantityRequest) (*emptypb.Empty, error) {
	itemId, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	if err = h.service.SubItemQuantity(ctx, itemId, req.GetQuantity()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ItemHandler) LockQuantity(ctx context.Context, req *api.LockQuantityRequest) (*emptypb.Empty, error) {
	itemId, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}
	if err = h.service.LockItemQuantity(ctx, itemId, req.GetQuantity()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ItemHandler) UnlockQuantity(ctx context.Context, req *api.UnlockQuantityRequest) (*emptypb.Empty, error) {
	itemId, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	if err = h.service.UnlockItemQuantity(ctx, itemId, req.GetQuantity()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ItemHandler) SubLockedQuantity(ctx context.Context, req *api.SubQuantityRequest) (*emptypb.Empty, error) {
	itemId, err := parseUUID(req.GetId())
	if err != nil {
		return nil, err
	}

	if err = h.service.SubLockedItemQuantity(ctx, itemId, req.GetQuantity()); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func parseUUID(id string) (uuid.UUID, error) {
	out, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, status.Error(codes.InvalidArgument, "invalid uuid")
	}

	return out, nil
}
