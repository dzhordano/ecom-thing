package grpc_server

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse ids",
		trace.WithAttributes(
			attribute.String("order_id", req.Order.GetId()),
			attribute.String("user_id", req.Order.GetUserId()),
		),
	)

	orderId, err := uuid.Parse(req.Order.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order uuid")
	}

	userId, err := uuid.Parse(req.Order.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user uuid")
	}

	span.AddEvent("call service")

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

	span.AddEvent("payment created")

	return &api.CreatePaymentResponse{
		Id: p.ID.String(),
	}, nil
}

// If user or admin wants to get payment status
func (h *PaymentHandler) GetPaymentStatus(ctx context.Context, req *api.GetPaymentStatusRequest) (*api.GetPaymentStatusResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("payment_id", req.GetId()),
		),
	)

	paymentId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid payment uuid")
	}

	userId, err := parseUUIDfromCtx(ctx)
	if err != nil {
		return nil, err
	}

	span.AddEvent("call service",
		trace.WithAttributes(
			attribute.String("user_id", userId.String()),
		),
	)

	p, err := h.service.GetPaymentStatus(ctx, paymentId, userId)
	if err != nil {
		return nil, err
	}

	span.AddEvent("payment status received")

	return &api.GetPaymentStatusResponse{
		Status: p,
	}, nil
}

// Say payment failed - canceled or expired for example, and needs to be retried
func (h *PaymentHandler) RetryPayment(ctx context.Context, req *api.RetryPaymentRequest) (*api.RetryPaymentResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("payment_id", req.GetId()),
		),
	)

	paymentId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid payment uuid")
	}

	userId, err := parseUUIDfromCtx(ctx)
	if err != nil {
		return nil, err
	}

	span.AddEvent("call service",
		trace.WithAttributes(
			attribute.String("user_id", userId.String()),
		),
	)

	if err := h.service.RetryPayment(ctx, paymentId, userId); err != nil {
		return nil, err
	}

	span.AddEvent("payment retried")

	return &api.RetryPaymentResponse{}, nil
}

// If user or admin wants to cancel payment
func (h *PaymentHandler) CancelPayment(ctx context.Context, req *api.CancelPaymentRequest) (*api.CancelPaymentResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("payment_id", req.GetId()),
		),
	)

	paymentId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid payment uuid")
	}

	userId, err := parseUUIDfromCtx(ctx)
	if err != nil {
		return nil, err
	}

	span.AddEvent("call service",
		trace.WithAttributes(
			attribute.String("user_id", userId.String()),
		),
	)

	if err := h.service.CancelPayment(ctx, paymentId, userId); err != nil {
		return nil, err
	}

	span.AddEvent("payment canceled")

	return &api.CancelPaymentResponse{}, nil
}

// User sends money (not with a card apparently, but just a transfer) so after payment is confirmed
func (h *PaymentHandler) ConfirmPayment(ctx context.Context, req *api.ConfirmPaymentRequest) (*api.ConfirmPaymentResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	span.AddEvent("parse id",
		trace.WithAttributes(
			attribute.String("payment_id", req.GetId()),
		),
	)

	paymentId, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid payment uuid")
	}

	userId, err := parseUUIDfromCtx(ctx)
	if err != nil {
		return nil, err
	}

	span.AddEvent("call service",
		trace.WithAttributes(
			attribute.String("user_id", userId.String()),
		),
	)

	if err := h.service.ConfirmPayment(ctx, paymentId, userId); err != nil {
		return nil, err
	}

	span.AddEvent("payment confirmed")

	return &api.ConfirmPaymentResponse{}, nil
}

func parseUUIDfromCtx(ctx context.Context) (uuid.UUID, error) {
	userIdStr, ok := ctx.Value("userId").(string)
	if !ok {
		return uuid.UUID{}, status.Error(codes.InvalidArgument, "no user uuid in context")
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		return uuid.UUID{}, status.Error(codes.InvalidArgument, "invalid user uuid")
	}

	return userId, nil
}
