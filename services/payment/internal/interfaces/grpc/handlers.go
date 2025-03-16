package grpc_server

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/payment/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/payment/internal/application/interfaces"
	api "github.com/dzhordano/ecom-thing/services/payment/pkg/api/payment/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PaymentHandler struct {
	api.UnimplementedPaymentServiceServer
	service interfaces.PaymentService
}

func NewPaymentHandler(service interfaces.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		service: service,
	}
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *api.CreatePaymentRequest) (*api.CreatePaymentResponse, error) {
	orderId, err := parseUUID(req.Order.GetId())
	if err != nil {
		return nil, err
	}

	userId, err := parseUUID(req.Order.GetUserId())
	if err != nil {
		return nil, err
	}

	p, err := h.service.CreatePayment(ctx, dto.CreatePaymentRequest{
		OrderId:       orderId,
		UserId:        userId,
		Currency:      req.Order.GetCurrency(),
		TotalPrice:    req.Order.GetTotalPrice(),
		PaymentMethod: req.GetPaymentMethod(),
		Description:   req.GetPaymentDescription(),
		RedirectURL:   req.GetRedirectUrl(),
	})
	if err != nil {
		return nil, err
	}

	return &api.CreatePaymentResponse{
		PaymentId: p.ID.String(),
	}, nil
}

// If user or admin wants to get payment status
func (h *PaymentHandler) GetPaymentStatus(ctx context.Context, req *api.GetPaymentStatusRequest) (*api.GetPaymentStatusResponse, error) {
	orderId, err := parseUUID(req.GetPaymentId())
	if err != nil {
		return nil, err
	}

	userId, err := parseUUIDfromCtx(ctx)
	if err != nil {
		return nil, err
	}

	p, err := h.service.GetPaymentStatus(ctx, orderId, userId)
	if err != nil {
		return nil, err
	}

	return &api.GetPaymentStatusResponse{
		Status: p,
	}, nil
}

// Say payment failed - canceled or expired for example, and needs to be retried
func (h *PaymentHandler) RetryPayment(ctx context.Context, req *api.RetryPaymentRequest) (*api.RetryPaymentResponse, error) {
	orderId, err := parseUUID(req.GetPaymentId())
	if err != nil {
		return nil, err
	}

	userId, err := parseUUIDfromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.RetryPayment(ctx, orderId, userId); err != nil {
		return nil, err
	}

	return &api.RetryPaymentResponse{}, nil
}

// If user or admin wants to cancel payment
func (h *PaymentHandler) CancelPayment(ctx context.Context, req *api.CancelPaymentRequest) (*api.CancelPaymentResponse, error) {
	orderId, err := parseUUID(req.GetPaymentId())
	if err != nil {
		return nil, err
	}

	userId, err := parseUUIDfromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.CancelPayment(ctx, orderId, userId); err != nil {
		return nil, err
	}

	return &api.CancelPaymentResponse{}, nil
}

// User sends money (not with a card apparently, but just a transfer) so after payment is confirmed
func (h *PaymentHandler) ConfirmPayment(ctx context.Context, req *api.ConfirmPaymentRequest) (*api.ConfirmPaymentResponse, error) {
	orderId, err := parseUUID(req.GetPaymentId())
	if err != nil {
		return nil, err
	}

	userId, err := parseUUIDfromCtx(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.service.ConfirmPayment(ctx, orderId, userId); err != nil {
		return nil, err
	}

	return &api.ConfirmPaymentResponse{}, nil
}

func parseUUID(id string) (uuid.UUID, error) {
	out, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, status.Error(codes.InvalidArgument, "invalid uuid")
	}

	return out, nil
}

func parseUUIDfromCtx(ctx context.Context) (uuid.UUID, error) {
	userIdStr, ok := ctx.Value("userId").(string)
	if !ok {
		return uuid.UUID{}, status.Error(codes.InvalidArgument, "invalid uuid")
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return uuid.UUID{}, status.Error(codes.InvalidArgument, "invalid uuid")
	}

	return userId, nil
}
