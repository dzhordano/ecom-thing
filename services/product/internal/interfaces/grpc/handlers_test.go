package grpc

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	mock_interfaces "github.com/dzhordano/ecom-thing/services/product/internal/interfaces/grpc/mocks"
	api "github.com/dzhordano/ecom-thing/services/product/pkg/grpc/product/v1"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestProductHandler_CreateProduct(t *testing.T) {
	type mockFunc func(s *mock_interfaces.MockProductService, req *api.CreateProductRequest)

	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	generatedUUID := uuid.New()
	testCtx := context.Background()

	tests := []struct {
		name         string
		ctx          context.Context
		req          *api.CreateProductRequest
		mockFunc     mockFunc
		expectedResp *api.CreateProductResponse
		expectedErr  error
	}{
		{
			name: "OK",
			ctx:  testCtx,
			req: &api.CreateProductRequest{
				Name:     "test",
				Category: "test",
				Desc:     "test",
				Price:    25.05,
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.CreateProductRequest) {
				s.EXPECT().CreateProduct(
					gomock.Any(),
					gomock.Eq(req.GetName()),
					gomock.Eq(req.GetDesc()),
					gomock.Eq(req.GetCategory()),
					gomock.Eq(req.GetPrice()),
				).Return(&domain.Product{
					ID:        generatedUUID,
					Name:      req.Name,
					Desc:      req.Desc,
					Category:  req.Category,
					Price:     req.Price,
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				}, nil).Times(1)
			},
			expectedResp: &api.CreateProductResponse{
				Product: &api.Product{
					Id:        generatedUUID.String(),
					Category:  "test",
					Name:      "test",
					Desc:      "test",
					Price:     25.05,
					CreatedAt: timestamppb.New(fixedTime),
					UpdatedAt: timestamppb.New(fixedTime),
				},
			},
			expectedErr: nil,
		},
		{
			name: "Invalid Argument",
			ctx:  testCtx,
			req: &api.CreateProductRequest{
				Name:     "",
				Category: "test",
				Desc:     "test",
				Price:    25.05,
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.CreateProductRequest) {
				s.EXPECT().CreateProduct(
					gomock.Any(),
					gomock.Eq(req.GetName()),
					gomock.Eq(req.GetDesc()),
					gomock.Eq(req.GetCategory()),
					gomock.Eq(req.GetPrice()),
				).Return(nil, domain.ErrInvalidArgument).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrInvalidArgument,
		},
		{
			"Invalid Price",
			testCtx,
			&api.CreateProductRequest{
				Name:     "test",
				Category: "test",
				Desc:     "test",
				Price:    math.Inf(1),
			},
			func(s *mock_interfaces.MockProductService, req *api.CreateProductRequest) {
				s.EXPECT().CreateProduct(
					gomock.Any(),
					gomock.Eq(req.GetName()),
					gomock.Eq(req.GetDesc()),
					gomock.Eq(req.GetCategory()),
					gomock.Eq(req.GetPrice()),
				).Return(nil, domain.ErrInvalidArgument).Times(1)
			},
			nil,
			domain.ErrInvalidArgument,
		},
		{
			name: "Product Already Exists",
			ctx:  testCtx,
			req: &api.CreateProductRequest{
				Name:     "test",
				Category: "test",
				Desc:     "test",
				Price:    25.05,
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.CreateProductRequest) {
				s.EXPECT().CreateProduct(
					gomock.Any(),
					gomock.Eq(req.GetName()),
					gomock.Eq(req.GetDesc()),
					gomock.Eq(req.GetCategory()),
					gomock.Eq(req.GetPrice()),
				).Return(nil, domain.ErrProductAlreadyExists).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrProductAlreadyExists,
		},
		{
			name: "Too Long Name",
			ctx:  testCtx,
			req: &api.CreateProductRequest{
				Name:     strings.Repeat("a", 257),
				Category: "test",
				Desc:     "test",
				Price:    1.5,
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.CreateProductRequest) {
				s.EXPECT().CreateProduct(
					gomock.Any(),
					gomock.Eq(req.GetName()),
					gomock.Eq(req.GetDesc()),
					gomock.Eq(req.GetCategory()),
					gomock.Eq(req.GetPrice()),
				).Return(nil, domain.ErrInvalidArgument).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrInvalidArgument,
		},
		{
			name: "Internal Error",
			ctx:  testCtx,
			req: &api.CreateProductRequest{
				Name:     "test",
				Category: "test",
				Desc:     "test",
				Price:    25.05,
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.CreateProductRequest) {
				s.EXPECT().CreateProduct(
					gomock.Any(),
					gomock.Eq(req.GetName()),
					gomock.Eq(req.GetDesc()),
					gomock.Eq(req.GetCategory()),
					gomock.Eq(req.GetPrice()),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			s := mock_interfaces.NewMockProductService(c)

			tt.mockFunc(s, tt.req)

			ctrl := NewProductHandler(s)

			resp, err := ctrl.CreateProduct(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.True(t, proto.Equal(tt.expectedResp, resp),
					"Expected:\n%v\nActual:\n%v", tt.expectedResp, resp)
			}
		})
	}
}

func TestProductHandler_GetProduct(t *testing.T) {
	type mockFunc func(s *mock_interfaces.MockProductService, req *api.GetProductRequest)

	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	generatedUUID := uuid.New()
	testCtx := context.Background()

	tests := []struct {
		name         string
		ctx          context.Context
		req          *api.GetProductRequest
		mockFunc     mockFunc
		expectedResp *api.GetProductResponse
		expectedErr  error
	}{
		{
			name: "OK",
			ctx:  testCtx,
			req: &api.GetProductRequest{
				Id: generatedUUID.String(),
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.GetProductRequest) {
				s.EXPECT().GetById(
					gomock.Any(),
					gomock.Eq(generatedUUID),
				).Return(&domain.Product{
					ID:        generatedUUID,
					Name:      "test",
					Desc:      "test",
					Category:  "test",
					Price:     25.05,
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				}, nil).Times(1)
			},
			expectedResp: &api.GetProductResponse{
				Product: &api.Product{
					Id:        generatedUUID.String(),
					Category:  "test",
					Name:      "test",
					Desc:      "test",
					Price:     25.05,
					CreatedAt: timestamppb.New(fixedTime),
					UpdatedAt: timestamppb.New(fixedTime),
				},
			},
			expectedErr: nil,
		},
		{
			name: "Invalid Product ID",
			ctx:  testCtx,
			req: &api.GetProductRequest{
				Id: "invalid",
			},
			mockFunc:     func(s *mock_interfaces.MockProductService, req *api.GetProductRequest) {},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid grpc id"),
		},
		{
			name: "Internal Error",
			ctx:  testCtx,
			req: &api.GetProductRequest{
				Id: generatedUUID.String(),
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.GetProductRequest) {
				s.EXPECT().GetById(
					gomock.Any(),
					gomock.Eq(generatedUUID),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			s := mock_interfaces.NewMockProductService(c)

			tt.mockFunc(s, tt.req)

			ctrl := NewProductHandler(s)

			resp, err := ctrl.GetProduct(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.True(t, proto.Equal(tt.expectedResp, resp),
					"Expected:\n%v\nActual:\n%v", tt.expectedResp, resp)
			}
		})
	}
}

func ptrVal[T any](t *testing.T, val T) *T {
	t.Helper()
	return &val
}

func TestProductHandler_SearchProducts(t *testing.T) {
	type mockFunc func(s *mock_interfaces.MockProductService, req *api.SearchProductsRequest)

	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	generatedUUID := uuid.New()
	testCtx := context.Background()

	tests := []struct {
		name         string
		ctx          context.Context
		req          *api.SearchProductsRequest
		mockFunc     mockFunc
		expectedResp *api.SearchProductsResponse
		expectedErr  error
	}{
		{
			name: "OK",
			ctx:  testCtx,
			req: &api.SearchProductsRequest{
				Query:    ptrVal(t, "test"),
				Category: ptrVal(t, "test"),
				MinPrice: ptrVal(t, 25.05),
				MaxPrice: ptrVal(t, 25.05),
				Limit:    10,
				Offset:   0,
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.SearchProductsRequest) {
				s.EXPECT().SearchProducts(
					gomock.Any(),
					map[string]any{
						"query":    req.Query,
						"category": req.Category,
						"minPrice": req.MinPrice,
						"maxPrice": req.MaxPrice,
					},
					gomock.Eq(req.GetLimit()),
					gomock.Eq(req.GetOffset()),
				).Return([]*domain.Product{
					{
						ID:        generatedUUID,
						Name:      "test",
						Desc:      "test",
						Category:  "test",
						Price:     25.05,
						CreatedAt: fixedTime,
						UpdatedAt: fixedTime,
					},
				}, nil).Times(1)
			},
			expectedResp: &api.SearchProductsResponse{
				Products: []*api.Product{
					{
						Id:        generatedUUID.String(),
						Category:  "test",
						Name:      "test",
						Desc:      "test",
						Price:     25.05,
						CreatedAt: timestamppb.New(fixedTime),
						UpdatedAt: timestamppb.New(fixedTime),
					},
				},
			},
			expectedErr: nil,
		},
		// TODO: add more tests
		{
			name: "Internal Error",
			ctx:  testCtx,
			req: &api.SearchProductsRequest{
				Query:    ptrVal(t, "test"),
				Category: ptrVal(t, "test"),
				MinPrice: ptrVal(t, 25.05),
				MaxPrice: ptrVal(t, 25.05),
				Limit:    10,
				Offset:   0,
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.SearchProductsRequest) {
				s.EXPECT().SearchProducts(
					gomock.Any(),
					map[string]any{
						"query":    req.Query,
						"category": req.Category,
						"minPrice": req.MinPrice,
						"maxPrice": req.MaxPrice,
					},
					gomock.Eq(req.GetLimit()),
					gomock.Eq(req.GetOffset()),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			s := mock_interfaces.NewMockProductService(c)

			tt.mockFunc(s, tt.req)

			ctrl := NewProductHandler(s)

			resp, err := ctrl.SearchProducts(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.True(t, proto.Equal(tt.expectedResp, resp),
					"Expected:\n%v\nActual:\n%v", tt.expectedResp, resp)
			}
		})
	}
}

func TestProductHandler_UpdateProduct(t *testing.T) {
	type mockFunc func(s *mock_interfaces.MockProductService, req *api.UpdateProductRequest)

	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	generatedUUID := uuid.New()
	testCtx := context.Background()

	tests := []struct {
		name         string
		ctx          context.Context
		req          *api.UpdateProductRequest
		mockFunc     mockFunc
		expectedResp *api.UpdateProductResponse
		expectedErr  error
	}{
		{
			name: "OK",
			ctx:  testCtx,
			req: &api.UpdateProductRequest{
				Id:       generatedUUID.String(),
				Name:     "test",
				Category: "test",
				Desc:     "test",
				IsActive: true,
				Price:    25.05,
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.UpdateProductRequest) {
				s.EXPECT().UpdateProduct(
					gomock.Any(),
					gomock.Eq(generatedUUID),
					gomock.Eq(req.GetName()),
					gomock.Eq(req.GetDesc()),
					gomock.Eq(req.GetCategory()),
					gomock.Eq(req.GetIsActive()),
					gomock.Eq(req.GetPrice()),
				).Return(&domain.Product{
					ID:        generatedUUID,
					Category:  "test",
					Name:      "test",
					Desc:      "test",
					Price:     25.05,
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				}, nil).Times(1)
			},
			expectedResp: &api.UpdateProductResponse{
				Product: &api.Product{
					Id:        generatedUUID.String(),
					Category:  "test",
					Name:      "test",
					Desc:      "test",
					Price:     25.05,
					CreatedAt: timestamppb.New(fixedTime),
					UpdatedAt: timestamppb.New(fixedTime),
				},
			},
			expectedErr: nil,
		},
		{
			name: "Invalid Product ID",
			ctx:  testCtx,
			req: &api.UpdateProductRequest{
				Id: "invalid",
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.UpdateProductRequest) {},
			expectedResp: &api.UpdateProductResponse{
				Product: nil,
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid grpc id"),
		},
		{
			name: "Product Not Found",
			ctx:  testCtx,
			req: &api.UpdateProductRequest{
				Id:       generatedUUID.String(),
				Name:     "test",
				Category: "test",
				Desc:     "test",
				Price:    25.05,
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.UpdateProductRequest) {
				s.EXPECT().UpdateProduct(
					gomock.Any(),
					gomock.Eq(generatedUUID),
					gomock.Eq(req.GetName()),
					gomock.Eq(req.GetDesc()),
					gomock.Eq(req.GetCategory()),
					gomock.Eq(req.GetIsActive()),
					gomock.Eq(req.GetPrice()),
				).Return(nil, domain.ErrProductNotFound).Times(1)
			},
			expectedResp: &api.UpdateProductResponse{
				Product: nil,
			},
			expectedErr: domain.ErrProductNotFound,
		},
		{
			name: "Internal Error",
			ctx:  testCtx,
			req: &api.UpdateProductRequest{
				Id:       generatedUUID.String(),
				Name:     "test",
				Category: "test",
				Desc:     "test",
				Price:    25.05,
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.UpdateProductRequest) {
				s.EXPECT().UpdateProduct(
					gomock.Any(),
					gomock.Eq(generatedUUID),
					gomock.Eq(req.GetName()),
					gomock.Eq(req.GetDesc()),
					gomock.Eq(req.GetCategory()),
					gomock.Eq(req.GetIsActive()),
					gomock.Eq(req.GetPrice()),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			s := mock_interfaces.NewMockProductService(c)

			tt.mockFunc(s, tt.req)

			ctrl := NewProductHandler(s)

			resp, err := ctrl.UpdateProduct(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.True(t, proto.Equal(tt.expectedResp, resp),
					"Expected:\n%v\nActual:\n%v", tt.expectedResp, resp)
			}
		})
	}
}

func TestProductHandler_DeactivateProduct(t *testing.T) {
	type mockFunc func(s *mock_interfaces.MockProductService, req *api.DeactivateProductRequest)

	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	generatedUUID := uuid.New()
	testCtx := context.Background()

	tests := []struct {
		name         string
		ctx          context.Context
		req          *api.DeactivateProductRequest
		mockFunc     mockFunc
		expectedResp *api.DeactivateProductResponse
		expectedErr  error
	}{
		{
			name: "OK",
			ctx:  testCtx,
			req: &api.DeactivateProductRequest{
				Id: generatedUUID.String(),
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.DeactivateProductRequest) {
				s.EXPECT().DeactivateProduct(
					gomock.Any(),
					gomock.Eq(generatedUUID),
				).Return(&domain.Product{
					ID:        generatedUUID,
					Name:      "test",
					Desc:      "test",
					Category:  "test",
					Price:     25.05,
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				}, nil).Times(1)
			},
			expectedResp: &api.DeactivateProductResponse{
				Product: &api.Product{
					Id:        generatedUUID.String(),
					Category:  "test",
					Name:      "test",
					Desc:      "test",
					Price:     25.05,
					CreatedAt: timestamppb.New(fixedTime),
					UpdatedAt: timestamppb.New(fixedTime),
				},
			},
			expectedErr: nil,
		},
		{
			name: "Invalid Product ID",
			ctx:  testCtx,
			req: &api.DeactivateProductRequest{
				Id: "invalid",
			},
			mockFunc:     func(s *mock_interfaces.MockProductService, req *api.DeactivateProductRequest) {},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid grpc id"),
		},
		{
			name: "Product Not Found",
			ctx:  testCtx,
			req: &api.DeactivateProductRequest{
				Id: generatedUUID.String(),
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.DeactivateProductRequest) {
				s.EXPECT().DeactivateProduct(
					gomock.Any(),
					gomock.Eq(generatedUUID),
				).Return(nil, domain.ErrProductNotFound).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrProductNotFound,
		},
		{
			name: "Internal Error",
			ctx:  testCtx,
			req: &api.DeactivateProductRequest{
				Id: generatedUUID.String(),
			},
			mockFunc: func(s *mock_interfaces.MockProductService, req *api.DeactivateProductRequest) {
				s.EXPECT().DeactivateProduct(
					gomock.Any(),
					gomock.Eq(generatedUUID),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			s := mock_interfaces.NewMockProductService(c)

			tt.mockFunc(s, tt.req)

			ctrl := NewProductHandler(s)

			resp, err := ctrl.DeactivateProduct(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.True(t, proto.Equal(tt.expectedResp, resp),
					"Expected:\n%v\nActual:\n%v", tt.expectedResp, resp)
			}
		})
	}
}
