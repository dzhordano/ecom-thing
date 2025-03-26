package service

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/payment/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/payment/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain/repository"
	"github.com/dzhordano/ecom-thing/services/payment/pkg/logger"
	"github.com/google/uuid"
)

type PaymentService struct {
	log  logger.Logger
	repo repository.PaymentRepository
}

func NewPaymerService(log logger.Logger, repo repository.PaymentRepository) interfaces.PaymentService {
	return &PaymentService{
		log:  log,
		repo: repo,
	}
}

// CreatePayment implements interfaces.PaymentService.
func (p *PaymentService) CreatePayment(ctx context.Context, req dto.CreatePaymentRequest) (*domain.Payment, error) {
	payment, err := domain.NewPayment(
		req.OrderId,
		req.UserId,
		req.Currency,
		req.TotalPrice,
		req.PaymentMethod,
		req.Description,
		req.RedirectURL,
		string(domain.PaymentPending),
	)
	if err != nil {
		p.log.Error("create payment error", "error", err)
		return nil, err
	}

	if err := payment.Validate(); err != nil {
		p.log.Error("create payment error", "error", err)
		return nil, err
	}

	if err = p.repo.Save(ctx, payment); err != nil {
		p.log.Error("create payment error", "error", err)
		return nil, err
	}

	p.log.Debug("create payment success")

	return payment, nil
}

// GetPaymentStatus implements interfaces.PaymentService.
func (p *PaymentService) GetPaymentStatus(ctx context.Context, paymentId, userId uuid.UUID) (string, error) {
	payment, err := p.repo.GetById(ctx, paymentId.String(), userId.String())
	if err != nil {
		p.log.Error("get payment status error", "error", err, "payment_id", paymentId.String())
		return "", err
	}

	p.log.Debug("get payment status success")

	return string(payment.Status), nil
}

// RetryPayment implements interfaces.PaymentService.
func (p *PaymentService) RetryPayment(ctx context.Context, paymentId, userId uuid.UUID) error {
	payment, err := p.repo.GetById(ctx, paymentId.String(), userId.String())
	if err != nil {
		p.log.Error("retry payment error", "error", err, "payment_id", paymentId.String())
		return err
	}

	if payment.Status == domain.PaymentCompleted {
		p.log.Error("retry payment error", "error", domain.ErrPaymentAlreadyCompleted, "payment_id", paymentId.String())
		return domain.ErrPaymentAlreadyCompleted
	}

	if payment.Status == domain.PaymentPending {
		p.log.Error("retry payment error", "error", domain.ErrPaymentAlreadyPending, "payment_id", paymentId.String())
		return domain.ErrPaymentAlreadyPending
	}

	payment.SetStatus(domain.PaymentPending)

	if err = p.repo.Update(ctx, payment); err != nil {
		p.log.Error("retry payment error", "error", err, "payment_id", paymentId.String())
		return err
	}

	p.log.Debug("retry payment success", "payment_id", paymentId.String())

	return nil
}

// CancelPayment implements interfaces.PaymentService.
func (p *PaymentService) CancelPayment(ctx context.Context, orderId, userId uuid.UUID) error {
	payment, err := p.repo.GetById(ctx, orderId.String(), userId.String())
	if err != nil {
		p.log.Error("cancel payment error", "error", err, "order_id", orderId.String())
		return err
	}

	if payment.Status != domain.PaymentPending {
		p.log.Error("cancel payment error", "error", domain.ErrInvalidPayment, "order_id", orderId.String())
		return domain.ErrInvalidPayment
	}

	payment.SetStatus(domain.PaymentCompleted)

	if err = p.repo.Update(ctx, payment); err != nil {
		p.log.Error("cancel payment error", "error", err, "order_id", orderId.String())
		return err
	}

	p.log.Debug("cancel payment success")

	return nil
}

// ConfirmPayment implements interfaces.PaymentService.
func (p *PaymentService) ConfirmPayment(ctx context.Context, orderId, userId uuid.UUID) error {
	payment, err := p.repo.GetById(ctx, orderId.String(), userId.String())
	if err != nil {
		p.log.Error("confirm payment error", "error", err, "order_id", orderId.String())
		return err
	}

	if payment.Status != domain.PaymentPending {
		p.log.Error("confirm payment error", "error", domain.ErrInvalidPayment, "order_id", orderId.String())
		return domain.ErrInvalidPayment
	}

	payment.SetStatus(domain.PaymentCompleted)

	if err = p.repo.Update(ctx, payment); err != nil {
		p.log.Error("confirm payment error", "error", err, "order_id", orderId.String())
		return err
	}

	p.log.Debug("confirm payment success")

	return nil
}
