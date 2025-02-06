package grpc

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/product/internal/interfaces/grpc/converter"
	api "github.com/dzhordano/ecom-thing/services/product/pkg/grpc/product/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductHandler struct {
	api.UnimplementedProductServiceV1Server
	service interfaces.ProductService
}

func NewProductHandler(service interfaces.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *api.CreateProductRequest) (*api.CreateProductResponse, error) {
	product, err := h.service.CreateProduct(ctx, req.GetName(), req.GetDesc(), req.GetCategory(), req.GetPrice())
	if err != nil {
		return nil, err
	}

	return &api.CreateProductResponse{
		Product: converter.ProductToProto(product),
	}, nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *api.UpdateProductRequest) (*api.UpdateProductResponse, error) {
	productId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid grpc id")
	}

	product, err := h.service.UpdateProduct(ctx, productId, req.GetName(), req.GetDesc(), req.GetCategory(), req.GetIsActive(), req.GetPrice())
	if err != nil {
		return nil, err
	}

	return &api.UpdateProductResponse{
		Product: converter.ProductToProto(product),
	}, nil
}

func (h *ProductHandler) DeactivateProduct(ctx context.Context, req *api.DeactivateProductRequest) (*api.DeactivateProductResponse, error) {
	productId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid grpc id")
	}

	product, err := h.service.DeactivateProduct(ctx, productId)
	if err != nil {
		return nil, err
	}

	return &api.DeactivateProductResponse{
		Product: converter.ProductToProto(product),
	}, nil
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *api.GetProductRequest) (*api.GetProductResponse, error) {
	productId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid grpc id")
	}

	product, err := h.service.GetById(ctx, productId)
	if err != nil {
		return nil, err
	}

	return &api.GetProductResponse{
		Product: converter.ProductToProto(product),
	}, nil
}

func (h *ProductHandler) SearchProducts(ctx context.Context, req *api.SearchProductsRequest) (*api.SearchProductsResponse, error) {
	products, err := h.service.SearchProducts(ctx, map[string]any{
		"query":    req.Query,
		"category": req.Category,
		"minPrice": req.MinPrice,
		"maxPrice": req.MaxPrice,
	}, req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, err
	}

	return &api.SearchProductsResponse{
		Products: converter.ManyProductsToProto(products),
	}, nil
}
