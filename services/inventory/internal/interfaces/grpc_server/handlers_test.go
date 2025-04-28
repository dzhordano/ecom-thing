package grpc_server

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	mock_interfaces "github.com/dzhordano/ecom-thing/services/inventory/internal/interfaces/grpc_server/mocks"
	api "github.com/dzhordano/ecom-thing/services/inventory/pkg/api/inventory/v1"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestItemHandler_GetItem(t *testing.T) {
	type mockBehavior func(s *mock_interfaces.MockItemService, id uuid.UUID)

	testId, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		req          *api.GetItemRequest
		mockBehavior mockBehavior
		expectedResp *api.GetItemResponse
		expectedErr  error
	}{
		{
			name: "OK",
			req: &api.GetItemRequest{
				ProductId: testId.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockItemService, id uuid.UUID) {
				s.EXPECT().GetItem(
					gomock.Any(),
					gomock.Eq(id),
				).Return(&domain.Item{
					ProductID:         testId,
					AvailableQuantity: 10,
					ReservedQuantity:  0,
				}, nil).Times(1)
			},
			expectedResp: &api.GetItemResponse{
				Item: &api.Item{
					ProductId:         testId.String(),
					AvailableQuantity: 10,
					ReservedQuantity:  0,
				},
			},
			expectedErr: nil,
		},
		{
			name: "ERROR",
			req: &api.GetItemRequest{
				ProductId: testId.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockItemService, id uuid.UUID) {
				s.EXPECT().GetItem(
					gomock.Any(),
					gomock.Eq(id),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockItemService := mock_interfaces.NewMockItemService(ctrl)
			tt.mockBehavior(mockItemService, testId)

			s := NewItemHandler(mockItemService)
			resp, err := s.GetItem(context.Background(), tt.req)

			if resp != nil || tt.expectedResp != nil {
				assert.Equal(t, tt.expectedResp.Item.ProductId, resp.Item.ProductId)
				assert.Equal(t, tt.expectedResp.Item.AvailableQuantity, resp.Item.AvailableQuantity)
				assert.Equal(t, tt.expectedResp.Item.ReservedQuantity, resp.Item.ReservedQuantity)
			} else {
				assert.Nil(t, resp)
			}
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestItemHandler_SetItem(t *testing.T) {
	type mockBehavior func(s *mock_interfaces.MockItemService, id uuid.UUID, quantity uint64, op string)

	testId, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		req          *api.SetItemRequest
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name: "OK",
			req: &api.SetItemRequest{
				Item: &api.ItemOP{
					ProductId: testId.String(),
					Quantity:  10,
				},
				OperationType: api.OperationType_OPERATION_TYPE_ADD,
			},
			mockBehavior: func(s *mock_interfaces.MockItemService, id uuid.UUID, quantity uint64, op string) {
				s.EXPECT().SetItemWithOp(
					gomock.Any(),
					gomock.Eq(id),
					gomock.Eq(quantity),
					gomock.Eq(op),
				).Return(nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name: "ERROR",
			req: &api.SetItemRequest{
				Item: &api.ItemOP{
					ProductId: testId.String(),
					Quantity:  10,
				},
				OperationType: api.OperationType_OPERATION_TYPE_ADD,
			},
			mockBehavior: func(s *mock_interfaces.MockItemService, id uuid.UUID, quantity uint64, op string) {
				s.EXPECT().SetItemWithOp(
					gomock.Any(),
					gomock.Eq(id),
					gomock.Eq(quantity),
					gomock.Eq(op),
				).Return(assert.AnError).Times(1)
			},
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockItemService := mock_interfaces.NewMockItemService(ctrl)
			tt.mockBehavior(mockItemService, testId, tt.req.Item.Quantity, protoOpToString(tt.req.OperationType))

			s := NewItemHandler(mockItemService)
			_, err := s.SetItem(context.Background(), tt.req)

			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestItemHandler_SetItems(t *testing.T) {
	type mockBehavior func(s *mock_interfaces.MockItemService, items map[string]uint64, op string)

	testId, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	testId2, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		req          *api.SetItemsRequest
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name: "OK",
			req: &api.SetItemsRequest{
				Items: []*api.ItemOP{
					{
						ProductId: testId.String(),
						Quantity:  10,
					},
					{
						ProductId: testId2.String(),
						Quantity:  10,
					},
				},
				OperationType: api.OperationType_OPERATION_TYPE_ADD,
			},
			mockBehavior: func(s *mock_interfaces.MockItemService, items map[string]uint64, op string) {
				s.EXPECT().SetItemsWithOp(
					gomock.Any(),
					gomock.Eq(items),
					gomock.Eq(op),
				).Return(nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name: "ERROR",
			req: &api.SetItemsRequest{
				Items: []*api.ItemOP{
					{
						ProductId: testId.String(),
						Quantity:  10,
					},
					{
						ProductId: testId2.String(),
						Quantity:  10,
					},
				},
				OperationType: api.OperationType_OPERATION_TYPE_ADD,
			},
			mockBehavior: func(s *mock_interfaces.MockItemService, items map[string]uint64, op string) {
				s.EXPECT().SetItemsWithOp(
					gomock.Any(),
					gomock.Eq(items),
					gomock.Eq(op),
				).Return(assert.AnError).Times(1)
			},
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			items := make(map[string]uint64)
			for _, item := range tt.req.Items {
				items[item.ProductId] = item.Quantity
			}

			mockItemService := mock_interfaces.NewMockItemService(ctrl)
			tt.mockBehavior(mockItemService, items, protoOpToString(tt.req.OperationType))

			s := NewItemHandler(mockItemService)
			_, err := s.SetItems(context.Background(), tt.req)

			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestItemHandler_IsReservable(t *testing.T) {
	type mockBehavior func(s *mock_interfaces.MockItemService, items map[string]uint64)

	testId, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	testId2, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		req          *api.IsReservableRequest
		mockBehavior mockBehavior
		expectedResp *api.IsReservableResponse
		expectedErr  error
	}{
		{
			name: "OK. quantity enough",
			req: &api.IsReservableRequest{
				Items: []*api.ItemOP{
					{
						ProductId: testId.String(),
						Quantity:  10,
					},
					{
						ProductId: testId2.String(),
						Quantity:  10,
					},
				},
			},
			mockBehavior: func(s *mock_interfaces.MockItemService, items map[string]uint64) {
				s.EXPECT().IsReservable(
					gomock.Any(),
					gomock.Eq(items),
				).Return(true, nil).Times(1)
			},
			expectedResp: &api.IsReservableResponse{
				IsReservable: true,
			},
		},
		{
			name: "OK. quantity not enough",
			req: &api.IsReservableRequest{
				Items: []*api.ItemOP{
					{
						ProductId: testId.String(),
						Quantity:  10,
					},
					{
						ProductId: testId2.String(),
						Quantity:  10,
					},
				},
			},
			mockBehavior: func(s *mock_interfaces.MockItemService, items map[string]uint64) {
				s.EXPECT().IsReservable(
					gomock.Any(),
					gomock.Eq(items),
				).Return(false, nil).Times(1)
			},
			expectedResp: &api.IsReservableResponse{
				IsReservable: false,
			},
		},
		{
			name: "ERROR",
			req: &api.IsReservableRequest{
				Items: []*api.ItemOP{
					{
						ProductId: testId.String(),
						Quantity:  10,
					},
					{
						ProductId: testId2.String(),
						Quantity:  10,
					},
				},
			},
			mockBehavior: func(s *mock_interfaces.MockItemService, items map[string]uint64) {
				s.EXPECT().IsReservable(
					gomock.Any(),
					gomock.Eq(items),
				).Return(false, assert.AnError).Times(1)
			},
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			items := make(map[string]uint64)
			for _, item := range tt.req.Items {
				items[item.ProductId] = item.Quantity
			}

			mockItemService := mock_interfaces.NewMockItemService(ctrl)
			tt.mockBehavior(mockItemService, items)

			s := NewItemHandler(mockItemService)
			resp, err := s.IsReservable(context.Background(), tt.req)

			if resp != nil || tt.expectedResp != nil {
				assert.Equal(t, tt.expectedResp, resp)
			} else {
				assert.Nil(t, resp)
			}
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func Test_parseUUID(t *testing.T) {
	testId, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		id      string
		want    uuid.UUID
		wantErr error
	}{
		{
			name: "OK",
			id:   testId.String(),
			want: testId,
		},
		{
			name:    "INVALID UUID",
			id:      "invalid uuid",
			wantErr: status.Error(codes.InvalidArgument, "invalid uuid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUUID(tt.id)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func Test_validUUID(t *testing.T) {
	testId, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name: "OK",
			id:   testId.String(),
		},
		{
			name:    "INVALID UUID",
			id:      "invalid uuid",
			wantErr: status.Error(codes.InvalidArgument, "invalid uuid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validUUID(tt.id)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func Test_protoOpToString(t *testing.T) {
	tests := []struct {
		name string
		req  api.OperationType
		want string
	}{
		{
			name: "OK",
			req:  api.OperationType_OPERATION_TYPE_ADD,
			want: domain.OperationAdd,
		},
		{
			name: "OK",
			req:  api.OperationType_OPERATION_TYPE_SUB,
			want: domain.OperationSub,
		},
		{
			name: "INVALID",
			req:  api.OperationType_OPERATION_TYPE_UNSPECIFIED,
			want: domain.OperationUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, protoOpToString(tt.req))
		})
	}
}
