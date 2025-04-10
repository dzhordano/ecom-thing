package grpc_server

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
	mock_interfaces "github.com/dzhordano/ecom-thing/services/payment/internal/interfaces/grpc_server/mocks"
	api "github.com/dzhordano/ecom-thing/services/payment/pkg/api/payment/v1"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"testing"
)

// TODO
func TestPaymentHandler_CreatePayment(t *testing.T) {
	t.Skip("todo")
}

func TestPaymentHandler_RetryPayment(t *testing.T) {
	type mockBehaviour func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID)

	testPaymentId := uuid.New()
	testUserId := uuid.New()

	tests := []struct {
		name          string
		ctx           context.Context
		req           *api.RetryPaymentRequest
		mockBehaviour mockBehaviour
		expectedResp  *api.RetryPaymentResponse
		expectedErr   error
	}{
		{
			name: "OK",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.RetryPaymentRequest{
				PaymentId: testPaymentId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {
				s.EXPECT().RetryPayment(
					gomock.Any(),
					gomock.Eq(paymentId),
					gomock.Eq(userId),
				).Return(nil).Times(1)
			},
			expectedResp: &api.RetryPaymentResponse{},
			expectedErr:  nil,
		},
		{
			name: "INVALID UUID",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.RetryPaymentRequest{
				PaymentId: "invalid uuid",
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.InvalidArgument, "invalid payment uuid"),
		},
		{
			name: "ERROR",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.RetryPaymentRequest{
				PaymentId: testPaymentId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {
				s.EXPECT().RetryPayment(
					gomock.Any(),
					gomock.Eq(paymentId),
					gomock.Eq(userId),
				).Return(assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			paymentService := mock_interfaces.NewMockPaymentService(ctrl)
			test.mockBehaviour(paymentService, testPaymentId, testUserId)

			h := NewPaymentHandler(paymentService)
			resp, err := h.RetryPayment(test.ctx, test.req)

			if test.expectedErr != nil {
				assert.Equal(t, test.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedResp, resp)
			}
		})
	}
}

// FIXME errors are hardcoded as in handlers
func TestPaymentHandler_GetPaymentStatus(t *testing.T) {
	type mockBehaviour func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID)

	testPaymentId := uuid.New()
	testUserId := uuid.New()

	tests := []struct {
		name          string
		ctx           context.Context
		req           *api.GetPaymentStatusRequest
		mockBehaviour mockBehaviour
		expectedResp  *api.GetPaymentStatusResponse
		expectedErr   error
	}{
		{
			name: "OK",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.GetPaymentStatusRequest{
				PaymentId: testPaymentId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {
				s.EXPECT().GetPaymentStatus(
					gomock.Any(),
					gomock.Eq(paymentId),
					gomock.Eq(userId),
				).Return(domain.PaymentPending.String(), nil).Times(1)
			},
			expectedResp: &api.GetPaymentStatusResponse{
				Status: domain.PaymentPending.String(),
			},
			expectedErr: nil,
		},
		{
			name: "INVALID UUID",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.GetPaymentStatusRequest{
				PaymentId: "invalid uuid",
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.InvalidArgument, "invalid payment uuid"),
		},
		{
			name: "INVALID USER UUID",
			ctx:  context.WithValue(context.Background(), "userId", "invalid uuid"),
			req: &api.GetPaymentStatusRequest{
				PaymentId: testPaymentId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.InvalidArgument, "invalid user uuid"),
		},
		{
			name: "ERROR",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.GetPaymentStatusRequest{
				PaymentId: testPaymentId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {
				s.EXPECT().GetPaymentStatus(
					gomock.Any(),
					gomock.Eq(paymentId),
					gomock.Eq(userId),
				).Return("", assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			paymentService := mock_interfaces.NewMockPaymentService(ctrl)
			tt.mockBehaviour(paymentService, testPaymentId, testUserId)

			paymentHandler := NewPaymentHandler(paymentService)

			resp, err := paymentHandler.GetPaymentStatus(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestPaymentHandler_CancelPayment(t *testing.T) {
	type mockBehaviour func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID)

	testPaymentId := uuid.New()
	testUserId := uuid.New()

	tests := []struct {
		name          string
		ctx           context.Context
		req           *api.CancelPaymentRequest
		mockBehaviour mockBehaviour
		expectedResp  *api.CancelPaymentResponse
		expectedErr   error
	}{
		{
			name: "OK",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.CancelPaymentRequest{
				PaymentId: testPaymentId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {
				s.EXPECT().CancelPayment(
					gomock.Any(),
					gomock.Eq(paymentId),
					gomock.Eq(userId),
				).Return(nil).Times(1)
			},
			expectedResp: &api.CancelPaymentResponse{},
			expectedErr:  nil,
		},
		{
			name: "INVALID UUID",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.CancelPaymentRequest{
				PaymentId: "invalid uuid",
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.InvalidArgument, "invalid payment uuid"),
		},
		{
			name: "INVALID USER UUID",
			ctx:  context.WithValue(context.Background(), "userId", "invalid uuid"),
			req: &api.CancelPaymentRequest{
				PaymentId: testPaymentId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.InvalidArgument, "invalid user uuid"),
		},
		{
			name: "ERROR",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.CancelPaymentRequest{
				PaymentId: testPaymentId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {
				s.EXPECT().CancelPayment(
					gomock.Any(),
					gomock.Eq(paymentId),
					gomock.Eq(userId),
				).Return(assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			paymentService := mock_interfaces.NewMockPaymentService(ctrl)
			tt.mockBehaviour(paymentService, testPaymentId, testUserId)

			paymentHandler := NewPaymentHandler(paymentService)

			resp, err := paymentHandler.CancelPayment(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestPaymentHandler_ConfirmPayment(t *testing.T) {
	type mockBehaviour func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID)

	testPaymentId := uuid.New()
	testUserId := uuid.New()

	tests := []struct {
		name          string
		ctx           context.Context
		req           *api.ConfirmPaymentRequest
		mockBehaviour mockBehaviour
		expectedResp  *api.ConfirmPaymentResponse
		expectedErr   error
	}{
		{
			name: "OK",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.ConfirmPaymentRequest{
				PaymentId: testPaymentId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {
				s.EXPECT().ConfirmPayment(
					gomock.Any(),
					gomock.Eq(paymentId),
					gomock.Eq(userId),
				).Return(nil).Times(1)
			},
			expectedResp: &api.ConfirmPaymentResponse{},
			expectedErr:  nil,
		},
		{
			name: "INVALID UUID",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.ConfirmPaymentRequest{
				PaymentId: "invalid uuid",
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.InvalidArgument, "invalid payment uuid"),
		},
		{
			name: "INVALID USER UUID",
			ctx:  context.WithValue(context.Background(), "userId", "invalid uuid"),
			req: &api.ConfirmPaymentRequest{
				PaymentId: testPaymentId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {},
			expectedResp:  nil,
			expectedErr:   status.Error(codes.InvalidArgument, "invalid user uuid"),
		},
		{
			name: "ERROR",
			ctx:  context.WithValue(context.Background(), "userId", testUserId.String()),
			req: &api.ConfirmPaymentRequest{
				PaymentId: testPaymentId.String(),
			},
			mockBehaviour: func(s *mock_interfaces.MockPaymentService, paymentId, userId uuid.UUID) {
				s.EXPECT().ConfirmPayment(
					gomock.Any(),
					gomock.Eq(paymentId),
					gomock.Eq(userId),
				).Return(assert.AnError).Times(1)
			},
			expectedResp: nil,
			expectedErr:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			paymentService := mock_interfaces.NewMockPaymentService(ctrl)
			tt.mockBehaviour(paymentService, testPaymentId, testUserId)

			paymentHandler := NewPaymentHandler(paymentService)

			resp, err := paymentHandler.ConfirmPayment(tt.ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func Test_parseUUIDfromCtx(t *testing.T) {
	testUserId := uuid.New()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		want    uuid.UUID
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				ctx: context.WithValue(context.Background(), "userId", testUserId.String()),
			},
			want:    testUserId,
			wantErr: false,
		},
		{
			name: "INVALID UUID",
			args: args{
				ctx: context.WithValue(context.Background(), "userId", "invalid uuid"),
			},
			want:    uuid.UUID{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUUIDfromCtx(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUUIDfromCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseUUIDfromCtx() got = %v, want %v", got, tt.want)
			}
		})
	}
}
