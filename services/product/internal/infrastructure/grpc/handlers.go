package grpc

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure/grpc/converter"
	api "github.com/dzhordano/ecom-thing/services/product/pkg/grpc/product/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (h *ProductHandler) GetProduct(ctx context.Context, req *api.GetProductRequest) (*api.GetProductResponse, error) {
	productId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}

	product, err := h.service.GetProduct(ctx, productId)
	if err != nil {
		return nil, err
	}

	return &api.GetProductResponse{
		Product: converter.ProductToProto(product),
	}, nil
}

func (h *ProductHandler) GetProducts(ctx context.Context, _ *emptypb.Empty) (*api.GetProductsResponse, error) {
	products, err := h.service.GetAllProducts(ctx)
	if err != nil {
		return nil, err
	}

	return &api.GetProductsResponse{
		Products: converter.ManyProductsToProto(products),
	}, nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *api.UpdateProductRequest) (*api.UpdateProductResponse, error) {
	productId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}

	product, err := h.service.UpdateProduct(ctx, productId, req.GetName(), req.GetDesc(), req.GetCategory(), req.GetPrice())
	if err != nil {
		return nil, err
	}

	return &api.UpdateProductResponse{
		Product: converter.ProductToProto(product),
	}, nil
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *api.DeleteProductRequest) (*api.DeleteProductResponse, error) {
	productId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product id")
	}

	product, err := h.service.DeleteProduct(ctx, productId)
	if err != nil {
		return nil, err
	}

	return &api.DeleteProductResponse{
		Product: converter.ProductToProto(product),
	}, nil
}
