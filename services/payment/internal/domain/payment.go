package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidArgument         = errors.New("invalid argument")
	ErrInvalidStatus           = errors.New("invalid status")
	ErrInvalidCurrency         = errors.New("invalid currency")
	ErrInvalidPaymentMethod    = errors.New("invalid payment method")
	ErrInvalidPayment          = errors.New("invalid payment") // Means that payment has invalid status for example
	ErrPaymentAlreadyPending   = errors.New("payment already pending")
	ErrPaymentAlreadyCompleted = errors.New("payment already completed")
	// Critical ones --->
	ErrPaymentCancelled = errors.New("payment cancelled") // FIXME надо ли?
	ErrPaymentFailed    = errors.New("payment failed")
	ErrInternal         = errors.New("internal error")
)

func CheckIfCriticalError(err error) bool {
	// No particular critical errors, so mark everything as critical expect for those we know are not critical.
	return !(errors.Is(err, ErrInvalidArgument) ||
		errors.Is(err, ErrInvalidStatus) ||
		errors.Is(err, ErrInvalidCurrency) ||
		errors.Is(err, ErrInvalidPaymentMethod) ||
		errors.Is(err, ErrInvalidPayment) ||
		errors.Is(err, ErrPaymentAlreadyPending) ||
		errors.Is(err, ErrPaymentAlreadyCompleted))
}

const (
	MaxPaymentDataLength = 255
)

type Payment struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	OrderID       uuid.UUID
	Currency      Currency
	TotalPrice    float64
	PaymentMethod PaymentMethod
	Description   string
	RedirectURL   string
	Status        Status
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Currency string

const (
	RUB Currency = "RUB"
	USD Currency = "USD"
	EUR Currency = "EUR"
)

var currencyMap = map[string]Currency{
	"RUB": RUB,
	"USD": USD,
	"EUR": EUR,
}

func NewCurrency(s string) (Currency, error) {
	if c, ok := currencyMap[s]; ok {
		return c, nil
	}
	return "", ErrInvalidCurrency
}

func (c Currency) String() string {
	return string(c)
}

type Status string

const (
	PaymentPending   Status = "pending"
	PaymentCompleted Status = "completed"
	PaymentCancelled Status = "cancelled"
	PaymentFailed    Status = "failed"
)

var statusMap = map[string]Status{
	"pending":   PaymentPending,
	"completed": PaymentCompleted,
	"cancelled": PaymentCancelled,
	"failed":    PaymentFailed,
}

func NewStatus(s string) (Status, error) {
	if st, ok := statusMap[s]; ok {
		return st, nil
	}
	return "", ErrInvalidStatus
}

func (s Status) String() string {
	return string(s)
}

type PaymentMethod string

const (
	PaymentMethodCard     PaymentMethod = "bank_card"
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodTransfer PaymentMethod = "transfer"
)

var paymentMethodMap = map[string]PaymentMethod{
	"bank_card": PaymentMethodCard,
	"cash":      PaymentMethodCash,
	"transfer":  PaymentMethodTransfer,
}

func NewPaymentMethod(s string) (PaymentMethod, error) {
	if pm, ok := paymentMethodMap[s]; ok {
		return pm, nil
	}
	return "", ErrInvalidPaymentMethod
}

func (s PaymentMethod) String() string {
	return string(s)
}

func NewPayment(orderId, userId uuid.UUID, currency string, totalPrice float64, paymentMethod, paymentDescription, redirectURL, status string) (*Payment, error) {
	paymentId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	c, err := NewCurrency(currency)
	if err != nil {
		return nil, err
	}

	ps, err := NewStatus(status)
	if err != nil {
		return nil, err
	}

	pm, err := NewPaymentMethod(paymentMethod)
	if err != nil {
		return nil, err
	}

	return &Payment{
		ID:            paymentId,
		UserID:        userId,
		OrderID:       orderId,
		Currency:      c,
		TotalPrice:    totalPrice,
		PaymentMethod: pm,
		Description:   paymentDescription,
		RedirectURL:   redirectURL,
		Status:        ps,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}, nil
}

func (p *Payment) IsValid() error {
	var errors []string

	if _, err := NewCurrency(p.Currency.String()); err != nil {
		errors = append(errors, err.Error())
	}

	if p.TotalPrice <= 0 {
		errors = append(errors, "total price must be greater than 0")
	}

	if _, err := NewPaymentMethod(p.PaymentMethod.String()); err != nil {
		errors = append(errors, err.Error())
	}

	if len(p.Description) > MaxPaymentDataLength {
		errors = append(errors, "payment data must be less than 256 characters")
	}

	if _, err := NewStatus(p.Status.String()); err != nil {
		errors = append(errors, err.Error())
	}

	if p.CreatedAt.After(p.UpdatedAt) {
		errors = append(errors, "created at must be before updated at")
	}

	if p.CreatedAt.After(time.Now().UTC()) {
		errors = append(errors, "created at must be before now")
	}

	if len(errors) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalidArgument, errors)
	}

	return nil
}

func (p *Payment) SetStatus(status Status) {
	p.Status = status
}

// MarkAsPaid updates the payment status to "paid".
func (p *Payment) MarkAsPaid() {
	p.SetStatus(PaymentCompleted)
}

// MarkAsCancelled updates the payment status to "cancelled" (if say user cancels the payment).
func (p *Payment) MarkAsCancelled() {
	p.SetStatus(PaymentCancelled)
}

// MarkAsFailed updates the payment status to "failed" (for example if it's expired).
func (p *Payment) MarkAsFailed() {
	p.SetStatus(PaymentFailed)
}
