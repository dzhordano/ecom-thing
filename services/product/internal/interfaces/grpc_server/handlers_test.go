package grpc_server

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	mock_interfaces "github.com/dzhordano/ecom-thing/services/product/internal/interfaces/grpc_server/mocks"
	api "github.com/dzhordano/ecom-thing/services/product/pkg/api/product/v1"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestProductHandler_CreateProduct(t *testing.T) {
	type mockBehaviour func(s *mock_interfaces.MockProductService, name, description, category string, price float64)

	respProduct := &domain.Product{
		ID:        uuid.New(),
		Name:      "test",
		Desc:      "cool test desc",
		Category:  "testing",
		Price:     15.25,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	tests := []struct {
		name          string
		req           *api.CreateProductRequest
		mockBehaviour mockBehaviour
		expectedResp  *api.CreateProductResponse
		expectedErr   error
	}{
		{
			name: "OK",
			req: &api.CreateProductRequest{
				Name:     respProduct.Name,
				Category: respProduct.Category,
				Desc:     respProduct.Desc,
				Price:    respProduct.Price,
			},
			mockBehaviour: func(s *mock_interfaces.MockProductService, name, description, category string, price float64) {
				s.EXPECT().CreateProduct(
					gomock.Any(),
					gomock.Eq(name),
					gomock.Eq(description),
					gomock.Eq(category),
					gomock.Eq(price),
				).Return(respProduct, nil).Times(1)
			},
			expectedResp: &api.CreateProductResponse{
				Product: &api.Product{
					Id:       respProduct.ID.String(),
					Name:     respProduct.Name,
					Desc:     respProduct.Desc,
					Category: respProduct.Category,
					IsActive: respProduct.IsActive,
					Price:    respProduct.Price,
					CreatedAt: &timestamppb.Timestamp{
						Seconds: respProduct.CreatedAt.Unix(),
						Nanos:   int32(respProduct.CreatedAt.Nanosecond()),
					},
					UpdatedAt: &timestamppb.Timestamp{
						Seconds: respProduct.UpdatedAt.Unix(),
						Nanos:   int32(respProduct.UpdatedAt.Nanosecond()),
					},
				},
			},
		},
		{
			name: "ERROR",
			req: &api.CreateProductRequest{
				Name:     respProduct.Name,
				Category: respProduct.Category,
				Desc:     respProduct.Desc,
				Price:    respProduct.Price,
			},
			mockBehaviour: func(s *mock_interfaces.MockProductService, name, description, category string, price float64) {
				s.EXPECT().CreateProduct(
					gomock.Any(),
					gomock.Eq(name),
					gomock.Eq(description),
					gomock.Eq(category),
					gomock.Eq(price),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(f *testing.T) {
			c := gomock.NewController(f)
			defer c.Finish()

			productService := mock_interfaces.NewMockProductService(c)
			tt.mockBehaviour(productService, tt.req.Name, tt.req.Desc, tt.req.Category, tt.req.Price)

			productHandler := NewProductHandler(productService)
			resp, err := productHandler.CreateProduct(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.ErrorIs(f, err, tt.expectedErr)
				assert.Nil(f, resp)
			} else {
				assert.Equal(f, resp, tt.expectedResp)
				assert.NoError(f, err)
			}
		})
	}
}

func TestProductHandler_DeactivateProduct(t *testing.T) {
	type mockBehaviour func(s *mock_interfaces.MockProductService, productId uuid.UUID)

	testProductId := uuid.New()

	timeNow := time.Now()
	timestamppbNow := timestamppb.New(timeNow)

	tests := []struct {
		name          string
		req           *api.DeactivateProductRequest
		mockBehaviour mockBehaviour
		expectedResp  *api.DeactivateProductResponse
		expectedErr   error
	}{
		{
			name: "OK",
			req: &api.DeactivateProductRequest{
				Id: testProductId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockProductService, productId uuid.UUID) {
				s.EXPECT().DeactivateProduct(
					gomock.Any(),
					gomock.Eq(productId),
				).Return(&domain.Product{
					ID:        productId,
					Name:      "test",
					Desc:      "test",
					Category:  "test",
					Price:     10,
					IsActive:  false,
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				}, nil).Times(1)
			},
			expectedResp: &api.DeactivateProductResponse{
				Product: &api.Product{
					Id:        testProductId.String(),
					Category:  "test",
					Name:      "test",
					Desc:      "test",
					Price:     10,
					IsActive:  false,
					CreatedAt: timestamppbNow,
					UpdatedAt: timestamppbNow,
				},
			},
			expectedErr: nil,
		},
		{
			name: "INVALID UUID",
			req: &api.DeactivateProductRequest{
				Id: "invalid uuid",
			},
			mockBehaviour: func(s *mock_interfaces.MockProductService, productId uuid.UUID) {},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.InvalidArgument, "invalid product id"),
		},
		{
			name: "ERROR",
			req: &api.DeactivateProductRequest{
				Id: testProductId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockProductService, productId uuid.UUID) {
				s.EXPECT().DeactivateProduct(
					gomock.Any(),
					gomock.Eq(productId),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(f *testing.T) {
			c := gomock.NewController(f)
			defer c.Finish()

			productService := mock_interfaces.NewMockProductService(c)
			tt.mockBehaviour(productService, testProductId)

			productHandler := NewProductHandler(productService)
			resp, err := productHandler.DeactivateProduct(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.ErrorIs(f, err, tt.expectedErr)
				assert.Nil(f, resp)
			} else {
				assert.Equal(f, resp.Product, tt.expectedResp.Product)
				assert.NoError(f, err)
			}
		})
	}
}

func TestProductHandler_GetProduct(t *testing.T) {
	type mockBehaviour func(s *mock_interfaces.MockProductService, productId uuid.UUID)

	testProductId := uuid.New()

	timeNow := time.Now()
	timestamppbNow := timestamppb.New(timeNow)

	tests := []struct {
		name          string
		req           *api.GetProductRequest
		mockBehaviour mockBehaviour
		expectedResp  *api.GetProductResponse
		expectedErr   error
	}{
		{
			name: "OK",
			req: &api.GetProductRequest{
				Id: testProductId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockProductService, productId uuid.UUID) {
				s.EXPECT().GetById(
					gomock.Any(),
					gomock.Eq(productId),
				).Return(&domain.Product{
					ID:        productId,
					Name:      "test",
					Desc:      "test",
					Category:  "test",
					Price:     10,
					IsActive:  true,
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				}, nil).Times(1)
			},
			expectedResp: &api.GetProductResponse{
				Product: &api.Product{
					Id:        testProductId.String(),
					Category:  "test",
					Name:      "test",
					Desc:      "test",
					Price:     10,
					IsActive:  true,
					CreatedAt: timestamppbNow,
					UpdatedAt: timestamppbNow,
				},
			},
			expectedErr: nil,
		},
		{
			name: "INVALID UUID",
			req: &api.GetProductRequest{
				Id: "invalid uuid",
			},
			mockBehaviour: func(s *mock_interfaces.MockProductService, productId uuid.UUID) {},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.InvalidArgument, "invalid product id"),
		},
		{
			name: "ERROR",
			req: &api.GetProductRequest{
				Id: testProductId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockProductService, productId uuid.UUID) {
				s.EXPECT().GetById(
					gomock.Any(),
					gomock.Eq(productId),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(f *testing.T) {
			c := gomock.NewController(f)
			defer c.Finish()

			productService := mock_interfaces.NewMockProductService(c)
			tt.mockBehaviour(productService, testProductId)

			productHandler := NewProductHandler(productService)
			resp, err := productHandler.GetProduct(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.ErrorIs(f, err, tt.expectedErr)
				assert.Nil(f, resp)
			} else {
				assert.Equal(f, resp.Product, tt.expectedResp.Product)
				assert.NoError(f, err)
			}
		})
	}
}

func TestProductHandler_SearchProducts(t *testing.T) {
	t.Skip("todo")
}

func TestProductHandler_UpdateProduct(t *testing.T) {
	type mockBehaviour func(
		s *mock_interfaces.MockProductService,
		productId uuid.UUID,
		name, description, category string,
		isActive bool,
		price float64,
	)

	testProductId := uuid.New()

	timeNow := time.Now()
	timestamppbNow := timestamppb.New(timeNow)

	tests := []struct {
		name          string
		req           *api.UpdateProductRequest
		mockBehaviour mockBehaviour
		expectedResp  *api.UpdateProductResponse
		expectedErr   error
	}{
		{
			name: "OK",
			req: &api.UpdateProductRequest{
				Id:       testProductId.String(),
				Name:     "test",
				Category: "test",
				Desc:     "test",
				Price:    10,
				IsActive: true,
			},
			mockBehaviour: func(s *mock_interfaces.MockProductService, productId uuid.UUID, name, description, category string, isActive bool, price float64) {
				s.EXPECT().UpdateProduct(
					gomock.Any(),
					gomock.Eq(productId),
					gomock.Eq(name),
					gomock.Eq(description),
					gomock.Eq(category),
					gomock.Eq(isActive),
					gomock.Eq(price),
				).Return(&domain.Product{
					ID:        productId,
					Name:      "test",
					Desc:      "test",
					Category:  "test",
					Price:     10,
					IsActive:  true,
					CreatedAt: timeNow,
					UpdatedAt: timeNow,
				}, nil).Times(1)
			},
			expectedResp: &api.UpdateProductResponse{
				Product: &api.Product{
					Id:        testProductId.String(),
					Category:  "test",
					Name:      "test",
					Desc:      "test",
					Price:     10,
					IsActive:  true,
					CreatedAt: timestamppbNow,
					UpdatedAt: timestamppbNow,
				},
			},
			expectedErr: nil,
		},
		{
			name: "INVALID UUID",
			req: &api.UpdateProductRequest{
				Id: "invalid uuid",
			},
			mockBehaviour: func(s *mock_interfaces.MockProductService, productId uuid.UUID, name, description, category string, isActive bool, price float64) {
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "invalid product id"),
		},
		{
			name: "ERROR",
			req: &api.UpdateProductRequest{
				Id: testProductId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockProductService, productId uuid.UUID, name, description, category string, isActive bool, price float64) {
				s.EXPECT().UpdateProduct(
					gomock.Any(),
					gomock.Eq(productId),
					gomock.Eq(name),
					gomock.Eq(description),
					gomock.Eq(category),
					gomock.Eq(isActive),
					gomock.Eq(price),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(f *testing.T) {
			c := gomock.NewController(f)
			defer c.Finish()

			productService := mock_interfaces.NewMockProductService(c)
			tt.mockBehaviour(productService, testProductId, tt.req.Name, tt.req.Desc, tt.req.Category, tt.req.IsActive, tt.req.Price)

			productHandler := NewProductHandler(productService)
			resp, err := productHandler.UpdateProduct(context.Background(), tt.req)

			if tt.expectedErr != nil {
				assert.ErrorIs(f, err, tt.expectedErr)
				assert.Nil(f, resp)
			} else {
				assert.Equal(f, resp, tt.expectedResp)
				assert.NoError(f, err)
			}
		})
	}
}
