package grpc_server

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/dzhordano/ecom-thing/services/product/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/product/internal/interfaces/grpc_server/converter"
	api "github.com/dzhordano/ecom-thing/services/product/pkg/api/product/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProductHandler struct {
	api.UnimplementedProductServiceServer
	service interfaces.ProductService
}

func NewProductHandler(service interfaces.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *api.CreateProductRequest) (*api.CreateProductResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("call service",
		trace.WithAttributes(
			attribute.Stringer("req", req),
		),
	)

	product, err := h.service.CreateProduct(ctx, req.GetName(), req.GetDesc(), req.GetCategory(), req.GetPrice())
	if err != nil {
		return nil, err
	}

	span.AddEvent("product created",
		trace.WithAttributes(
			attribute.Stringer("product_id", product.ID),
		),
	)

	return &api.CreateProductResponse{
		Product: converter.ProductToProto(product),
	}, nil
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *api.UpdateProductRequest) (*api.UpdateProductResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("product_id", req.GetId()),
		),
	)

	productId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid grpc_server id")
	}

	span.AddEvent("call service")

	product, err := h.service.UpdateProduct(ctx, productId, req.GetName(), req.GetDesc(), req.GetCategory(), req.GetIsActive(), req.GetPrice())
	if err != nil {
		return nil, err
	}

	span.AddEvent("product updated",
		trace.WithAttributes(
			attribute.Stringer("product_id", product.ID),
		),
	)

	return &api.UpdateProductResponse{
		Product: converter.ProductToProto(product),
	}, nil
}

func (h *ProductHandler) DeactivateProduct(ctx context.Context, req *api.DeactivateProductRequest) (*api.DeactivateProductResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("product_id", req.GetId()),
		),
	)

	productId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid grpc_server id")
	}

	span.AddEvent("call service")

	product, err := h.service.DeactivateProduct(ctx, productId)
	if err != nil {
		return nil, err
	}

	span.AddEvent("product deactivated",
		trace.WithAttributes(
			attribute.Stringer("product_id", product.ID),
		),
	)

	return &api.DeactivateProductResponse{
		Product: converter.ProductToProto(product),
	}, nil
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *api.GetProductRequest) (*api.GetProductResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("product_id", req.GetId()),
		),
	)

	productId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid grpc_server id")
	}

	span.AddEvent("call service")

	product, err := h.service.GetById(ctx, productId)
	if err != nil {
		return nil, err
	}

	span.AddEvent("product retrieved",
		trace.WithAttributes(
			attribute.Stringer("product_id", product.ID),
		),
	)

	return &api.GetProductResponse{
		Product: converter.ProductToProto(product),
	}, nil
}

func (h *ProductHandler) SearchProducts(ctx context.Context, req *api.SearchProductsRequest) (*api.SearchProductsResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("call service",
		trace.WithAttributes(
			attribute.Stringer("req", req),
		),
	)

	products, err := h.service.SearchProducts(ctx, map[string]any{
		"query":    req.Query,
		"category": req.Category,
		"minPrice": req.MinPrice,
		"maxPrice": req.MaxPrice,
		"limit":    req.Limit,
		"offset":   req.Offset,
	})
	if err != nil {
		return nil, err
	}

	span.AddEvent("products found",
		trace.WithAttributes(
			attribute.Int("count", len(products)),
		),
	)

	return &api.SearchProductsResponse{
		Products: converter.ManyProductsToProto(products),
	}, nil
}
