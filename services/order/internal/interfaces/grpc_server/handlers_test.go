package grpc_server

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	mock_interfaces "github.com/dzhordano/ecom-thing/services/order/internal/interfaces/grpc_server/mocks"
	api "github.com/dzhordano/ecom-thing/services/order/pkg/api/order/v1"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestItemHandler_CancelOrder(t *testing.T) {
	type mockBehavior func(s *mock_interfaces.MockOrderService, orderId uuid.UUID)

	testId, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		req          *api.CancelOrderRequest
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name: "OK",
			req: &api.CancelOrderRequest{
				OrderId: testId.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().CancelOrder(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name: "NOT FOUND",
			req: &api.CancelOrderRequest{
				OrderId: testId.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().CancelOrder(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(domain.ErrOrderNotFound).Times(1)
			},
			expectedErr: domain.ErrOrderNotFound,
		},
		{
			name: "INVALID UUID",
			req: &api.CancelOrderRequest{
				OrderId: "invalid",
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {},
			expectedErr:  domain.ErrInvalidUUID,
		},
		{
			name: "ALREADY COMPLETED",
			req: &api.CancelOrderRequest{
				OrderId: testId.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().CancelOrder(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(domain.ErrOrderAlreadyCompleted).Times(1)
			},
			expectedErr: domain.ErrOrderAlreadyCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockItemService := mock_interfaces.NewMockOrderService(ctrl)
			tt.mockBehavior(mockItemService, testId)

			s := NewOrderHandler(mockItemService)

			// No resp value. It returns empty struct.
			_, err := s.CancelOrder(context.Background(), tt.req)

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestItemHandler_CompleteOrder(t *testing.T) {
	type mockBehavior func(s *mock_interfaces.MockOrderService, orderId uuid.UUID)

	testId, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		req          *api.CompleteOrderRequest
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name: "OK",
			req: &api.CompleteOrderRequest{
				OrderId: testId.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().CompleteOrder(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name: "NOT FOUND",
			req: &api.CompleteOrderRequest{
				OrderId: testId.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().CompleteOrder(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(domain.ErrOrderNotFound).Times(1)
			},
			expectedErr: domain.ErrOrderNotFound,
		},
		{
			name: "INVALID UUID",
			req: &api.CompleteOrderRequest{
				OrderId: "invalid",
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {},
			expectedErr:  domain.ErrInvalidUUID,
		},
		{
			name: "ALREADY CANCELLED",
			req: &api.CompleteOrderRequest{
				OrderId: testId.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().CompleteOrder(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(domain.ErrOrderAlreadyCancelled).Times(1)
			},
			expectedErr: domain.ErrOrderAlreadyCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockItemService := mock_interfaces.NewMockOrderService(ctrl)
			tt.mockBehavior(mockItemService, testId)

			s := NewOrderHandler(mockItemService)

			// No resp value. It returns empty struct.
			_, err := s.CompleteOrder(context.Background(), tt.req)

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestItemHandler_GetOrder(t *testing.T) {
	type mockBehavior func(s *mock_interfaces.MockOrderService, orderId uuid.UUID)

	testOrder, err := domain.NewOrder(
		uuid.UUID{},
		"tt",
		domain.OrderPending.String(),
		domain.RUB.String(),
		10.00,
		1.00,
		domain.BankCard.String(),
		domain.Pickup.String(),
		"tt st.",
		time.Now().Add(time.Hour),
		[]domain.Item{
			{
				ProductID: uuid.UUID{},
				Quantity:  1,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	protoTestOrder := &api.Order{
		OrderId:         testOrder.ID.String(),
		UserId:          testOrder.UserID.String(),
		Description:     testOrder.Description,
		Status:          testOrder.Status.String(),
		Currency:        testOrder.Currency.String(),
		TotalPrice:      testOrder.TotalPrice,
		PaymentMethod:   testOrder.PaymentMethod.String(),
		DeliveryMethod:  testOrder.DeliveryMethod.String(),
		DeliveryAddress: testOrder.DeliveryAddress,
		DeliveryDate: &timestamppb.Timestamp{
			Seconds: testOrder.DeliveryDate.Unix(),
			Nanos:   int32(testOrder.DeliveryDate.Nanosecond()),
		},
		CreatedAt: &timestamppb.Timestamp{
			Seconds: testOrder.CreatedAt.Unix(),
			Nanos:   int32(testOrder.CreatedAt.Nanosecond()),
		},
		UpdatedAt: &timestamppb.Timestamp{
			Seconds: testOrder.UpdatedAt.Unix(),
			Nanos:   int32(testOrder.UpdatedAt.Nanosecond()),
		},
		Items: []*api.Item{
			{
				ItemId:   testOrder.Items[0].ProductID.String(),
				Quantity: testOrder.Items[0].Quantity,
			},
		},
	}

	tests := []struct {
		name         string
		req          *api.GetOrderRequest
		mockBehavior mockBehavior
		expectedResp *api.GetOrderResponse
		expectedErr  error
	}{
		{
			name: "OK",
			req: &api.GetOrderRequest{
				OrderId: testOrder.ID.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().GetById(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(testOrder, nil).Times(1)
			},
			expectedResp: &api.GetOrderResponse{
				Order: protoTestOrder,
			},
			expectedErr: nil,
		},
		{
			name: "ORDER NOT FOUND",
			req: &api.GetOrderRequest{
				OrderId: testOrder.ID.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().GetById(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(nil, domain.ErrOrderNotFound).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrOrderNotFound,
		},
		{
			name: "INVALID UUID",
			req: &api.GetOrderRequest{
				OrderId: "invalid uuid",
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {},
			expectedResp: nil,
			expectedErr:  domain.ErrInvalidUUID,
		},
		{
			name: "INTERNAL",
			req: &api.GetOrderRequest{
				OrderId: testOrder.ID.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().GetById(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrderService := mock_interfaces.NewMockOrderService(ctrl)
			tt.mockBehavior(mockOrderService, testOrder.ID)

			s := NewOrderHandler(mockOrderService)

			// No resp value. It returns empty struct.
			resp, err := s.GetOrder(context.Background(), tt.req)

			assert.Equal(t, tt.expectedResp, resp)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestItemHandler_CreateOrder(t *testing.T) {
	type mockBehavior func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest)

	testOrder := &domain.Order{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Description:     "test description",
		Status:          domain.OrderPending,
		Currency:        domain.RUB,
		TotalPrice:      100,
		PaymentMethod:   domain.Cash,
		DeliveryMethod:  domain.Pickup,
		DeliveryAddress: "test address",
		DeliveryDate:    time.Now().UTC(),
		Items: domain.Items{
			{
				ProductID: uuid.New(),
				Quantity:  1,
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	rpcTestOrder := &api.Order{
		OrderId:         testOrder.ID.String(),
		UserId:          testOrder.UserID.String(),
		Description:     testOrder.Description,
		Status:          testOrder.Status.String(),
		Currency:        testOrder.Currency.String(),
		TotalPrice:      testOrder.TotalPrice,
		PaymentMethod:   testOrder.PaymentMethod.String(),
		DeliveryMethod:  testOrder.DeliveryMethod.String(),
		DeliveryAddress: testOrder.DeliveryAddress,
		DeliveryDate: &timestamppb.Timestamp{
			Seconds: testOrder.DeliveryDate.Unix(),
			Nanos:   int32(testOrder.DeliveryDate.Nanosecond()),
		},
		Items: []*api.Item{
			{
				ItemId:   testOrder.Items[0].ProductID.String(),
				Quantity: testOrder.Items[0].Quantity,
			},
		},
		CreatedAt: &timestamppb.Timestamp{
			Seconds: testOrder.CreatedAt.Unix(),
			Nanos:   int32(testOrder.CreatedAt.Nanosecond()),
		},
		UpdatedAt: &timestamppb.Timestamp{
			Seconds: testOrder.UpdatedAt.Unix(),
			Nanos:   int32(testOrder.UpdatedAt.Nanosecond()),
		},
	}

	testCoupon := "test"

	rpcReq := &api.CreateOrderRequest{
		Description:     &rpcTestOrder.Description,
		Currency:        rpcTestOrder.Currency,
		Coupon:          &testCoupon,
		PaymentMethod:   rpcTestOrder.PaymentMethod,
		DeliveryMethod:  rpcTestOrder.DeliveryMethod,
		DeliveryAddress: rpcTestOrder.DeliveryAddress,
		DeliveryDate:    rpcTestOrder.DeliveryDate,
		Items:           rpcTestOrder.Items,
	}

	rpcReqInvalidItemUUID := &api.CreateOrderRequest{
		Description:     &rpcTestOrder.Description,
		Currency:        rpcTestOrder.Currency,
		Coupon:          &testCoupon,
		PaymentMethod:   rpcTestOrder.PaymentMethod,
		DeliveryMethod:  rpcTestOrder.DeliveryMethod,
		DeliveryAddress: rpcTestOrder.DeliveryAddress,
		DeliveryDate:    rpcTestOrder.DeliveryDate,
		Items: []*api.Item{
			{
				ItemId:   "invalid uuid",
				Quantity: 1,
			},
		},
	}

	svcReq := dto.CreateOrderRequest{
		Description:     testOrder.Description,
		Currency:        testOrder.Currency.String(),
		Coupon:          testCoupon,
		PaymentMethod:   testOrder.PaymentMethod.String(),
		DeliveryMethod:  testOrder.DeliveryMethod.String(),
		DeliveryAddress: testOrder.DeliveryAddress,
		DeliveryDate:    testOrder.DeliveryDate,
		Items:           testOrder.Items,
	}

	tests := []struct {
		name         string
		req          *api.CreateOrderRequest
		info         dto.CreateOrderRequest
		mockBehavior mockBehavior
		expectedResp *api.CreateOrderResponse
		expectedErr  error
	}{
		{
			name: "OK",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(testOrder, nil).Times(1)
			},
			expectedResp: &api.CreateOrderResponse{
				Order: &api.Order{
					OrderId:         testOrder.ID.String(),
					UserId:          testOrder.UserID.String(),
					Description:     testOrder.Description,
					Status:          testOrder.Status.String(),
					Currency:        testOrder.Currency.String(),
					TotalPrice:      testOrder.TotalPrice,
					PaymentMethod:   testOrder.PaymentMethod.String(),
					DeliveryMethod:  testOrder.DeliveryMethod.String(),
					DeliveryAddress: testOrder.DeliveryAddress,
					DeliveryDate: &timestamppb.Timestamp{
						Seconds: testOrder.DeliveryDate.Unix(),
						Nanos:   int32(testOrder.DeliveryDate.Nanosecond()),
					},
					Items: []*api.Item{
						{
							ItemId:   testOrder.Items[0].ProductID.String(),
							Quantity: testOrder.Items[0].Quantity,
						},
					},
					CreatedAt: &timestamppb.Timestamp{
						Seconds: testOrder.CreatedAt.Unix(),
						Nanos:   int32(testOrder.CreatedAt.Nanosecond()),
					},
					UpdatedAt: &timestamppb.Timestamp{
						Seconds: testOrder.UpdatedAt.Unix(),
						Nanos:   int32(testOrder.UpdatedAt.Nanosecond()),
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "COUPON NOT FOUND",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, domain.ErrCouponNotFound).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrCouponNotFound,
		},
		{
			name: "COUPON EXPIRED",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, domain.ErrCouponExpired).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrCouponExpired,
		},
		{
			name: "COUPON NOT ACTIVE",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, domain.ErrCouponNotActive).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrCouponNotActive,
		},
		{
			name: "PRODUCT UNAVAILABLE",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, domain.ErrProductUnavailable).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrProductUnavailable,
		},
		{
			name: "NOT ENOUGH QUANTITY",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, domain.ErrNotEnoughQuantity).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrNotEnoughQuantity,
		},
		{
			name: "INVENTORY UNAVAILABLE",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, domain.ErrInventoryUnavailable).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrInventoryUnavailable,
		},
		{
			name: "INVALID STATUS",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, domain.ErrInvalidOrderStatus).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrInvalidOrderStatus,
		},
		{
			name: "INVALID PAYMENT METHOD",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, domain.ErrInvalidPaymentMethod).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrInvalidPaymentMethod,
		},
		{
			name: "INVALID DELIVERY METHOD",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, domain.ErrInvalidDeliveryMethod).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrInvalidDeliveryMethod,
		},
		{
			name: "INVALID DELIVERY ADDRESS",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, domain.ErrInvalidDeliveryAddress).Times(1)
			},
			expectedResp: nil,
			expectedErr:  domain.ErrInvalidDeliveryAddress,
		},
		{
			name:         "INVALID UUID",
			req:          rpcReqInvalidItemUUID,
			info:         dto.CreateOrderRequest{},
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {},
			expectedResp: nil,
			expectedErr:  domain.ErrInvalidUUID,
		},
		{
			name: "INTERNAL",
			req:  rpcReq,
			info: svcReq,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.CreateOrderRequest) {
				s.EXPECT().CreateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrderService := mock_interfaces.NewMockOrderService(ctrl)
			tt.mockBehavior(mockOrderService, tt.info)

			s := NewOrderHandler(mockOrderService)

			resp, err := s.CreateOrder(context.Background(), tt.req)

			if resp != nil || tt.expectedResp != nil {
				assert.Equal(t, tt.expectedResp.Order, resp.Order)
			}
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestItemHandler_DeleteOrder(t *testing.T) {
	type mockBehavior func(s *mock_interfaces.MockOrderService, orderId uuid.UUID)

	testId, err := uuid.NewUUID()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		req          *api.DeleteOrderRequest
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name: "OK",
			req: &api.DeleteOrderRequest{
				OrderId: testId.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().DeleteOrder(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(nil).Times(1)
			},
			expectedErr: nil,
		},
		{
			name: "NOT FOUND",
			req: &api.DeleteOrderRequest{
				OrderId: testId.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().DeleteOrder(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(domain.ErrOrderNotFound).Times(1)
			},
			expectedErr: domain.ErrOrderNotFound,
		},
		{
			name: "INVALID UUID",
			req: &api.DeleteOrderRequest{
				OrderId: "invalid uuid",
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {},
			expectedErr:  domain.ErrInvalidUUID,
		},
		{
			name: "INTERNAL",
			req: &api.DeleteOrderRequest{
				OrderId: testId.String(),
			},
			mockBehavior: func(s *mock_interfaces.MockOrderService, orderId uuid.UUID) {
				s.EXPECT().DeleteOrder(
					gomock.Any(),
					gomock.Eq(orderId),
				).Return(assert.AnError).Times(1)
			},
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrderService := mock_interfaces.NewMockOrderService(ctrl)
			tt.mockBehavior(mockOrderService, testId)

			s := NewOrderHandler(mockOrderService)

			// No resp value. It returns empty struct.
			_, err := s.DeleteOrder(context.Background(), tt.req)

			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

// TODO
func TestItemHandler_ListOrders(t *testing.T) {
	t.Skip("not implemented")
}

// TODO
// This one doesn't do anything really, handler func just passes a map to service layer (e.t. mock in test).
func TestItemHandler_SearchOrders(t *testing.T) {
	t.Skip("not implemented")
}

func TestItemHandler_UpdateOrder(t *testing.T) {
	type mockBehavior func(s *mock_interfaces.MockOrderService, info dto.UpdateOrderRequest)

	testOrder := &domain.Order{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Description:     "test description",
		Status:          domain.OrderPending,
		Currency:        domain.RUB,
		TotalPrice:      100,
		PaymentMethod:   domain.Cash,
		DeliveryMethod:  domain.Pickup,
		DeliveryAddress: "test address",
		DeliveryDate:    time.Now().UTC(),
		Items: domain.Items{
			{
				ProductID: uuid.New(),
				Quantity:  1,
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	rpcTestOrder := &api.Order{
		OrderId:         testOrder.ID.String(),
		UserId:          testOrder.UserID.String(),
		Description:     testOrder.Description,
		Status:          testOrder.Status.String(),
		Currency:        testOrder.Currency.String(),
		TotalPrice:      testOrder.TotalPrice,
		PaymentMethod:   testOrder.PaymentMethod.String(),
		DeliveryMethod:  testOrder.DeliveryMethod.String(),
		DeliveryAddress: testOrder.DeliveryAddress,
		DeliveryDate: &timestamppb.Timestamp{
			Seconds: testOrder.DeliveryDate.Unix(),
			Nanos:   int32(testOrder.DeliveryDate.Nanosecond()),
		},
		Items: []*api.Item{
			{
				ItemId:   testOrder.Items[0].ProductID.String(),
				Quantity: testOrder.Items[0].Quantity,
			},
		},
		CreatedAt: &timestamppb.Timestamp{
			Seconds: testOrder.CreatedAt.Unix(),
			Nanos:   int32(testOrder.CreatedAt.Nanosecond()),
		},
		UpdatedAt: &timestamppb.Timestamp{
			Seconds: testOrder.UpdatedAt.Unix(),
			Nanos:   int32(testOrder.UpdatedAt.Nanosecond()),
		},
	}

	validInfo := dto.UpdateOrderRequest{
		OrderID:         testOrder.ID,
		Description:     &testOrder.Description,
		Status:          (*string)(&testOrder.Status),
		TotalPrice:      &testOrder.TotalPrice,
		PaymentMethod:   (*string)(&testOrder.PaymentMethod),
		DeliveryMethod:  (*string)(&testOrder.DeliveryMethod),
		DeliveryAddress: &testOrder.DeliveryAddress,
		DeliveryDate:    testOrder.DeliveryDate,
		Items:           testOrder.Items,
	}

	rpcReq := &api.UpdateOrderRequest{
		OrderId:         rpcTestOrder.OrderId,
		Description:     &rpcTestOrder.Description,
		Status:          &rpcTestOrder.Status,
		TotalPrice:      &rpcTestOrder.TotalPrice,
		PaymentMethod:   &rpcTestOrder.PaymentMethod,
		DeliveryMethod:  &rpcTestOrder.DeliveryMethod,
		DeliveryAddress: &rpcTestOrder.DeliveryAddress,
		DeliveryDate:    rpcTestOrder.DeliveryDate,
		Items:           rpcTestOrder.Items,
	}

	tests := []struct {
		name         string
		req          *api.UpdateOrderRequest
		info         dto.UpdateOrderRequest
		mockBehavior mockBehavior
		expectedResp *api.UpdateOrderResponse
		expectedErr  error
	}{
		{
			name: "OK",
			req:  rpcReq,
			info: validInfo,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.UpdateOrderRequest) {
				s.EXPECT().UpdateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(testOrder, nil).Times(1)
			},
			expectedResp: &api.UpdateOrderResponse{
				Order: rpcTestOrder,
			},
			expectedErr: nil,
		},
		{
			name: "INVALID UUID",
			req: &api.UpdateOrderRequest{
				OrderId: "invalid uuid",
			},
			info:         dto.UpdateOrderRequest{},
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.UpdateOrderRequest) {},
			expectedResp: nil,
			expectedErr:  domain.ErrInvalidUUID,
		},
		{
			name: "INTERNAL",
			req:  rpcReq,
			info: validInfo,
			mockBehavior: func(s *mock_interfaces.MockOrderService, info dto.UpdateOrderRequest) {
				s.EXPECT().UpdateOrder(
					gomock.Any(),
					gomock.Eq(info),
				).Return(nil, assert.AnError).Times(1)
			},
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockOrderService := mock_interfaces.NewMockOrderService(ctrl)
			tt.mockBehavior(mockOrderService, tt.info)

			s := NewOrderHandler(mockOrderService)

			resp, err := s.UpdateOrder(context.Background(), tt.req)

			assert.Equal(t, tt.expectedResp, resp)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func Test_timeFromProtoIfNotZero(t *testing.T) {
	tests := []struct {
		name string
		req  *timestamppb.Timestamp
		want time.Time
	}{
		{
			name: "OK",
			req: &timestamppb.Timestamp{
				Seconds: 1620000000,
				Nanos:   123456789,
			},
			want: time.Unix(1620000000, 123456789).UTC(),
		},
		{
			name: "OK. timestamppb zero value",
			req:  &timestamppb.Timestamp{},
			want: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, timeFromProtoIfNotZero(tt.req))
		})
	}
}
