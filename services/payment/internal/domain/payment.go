package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Service for payment
// FIXME ХЗ КУДА

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

// TODO Опять без констант, но норм т.к. не над излишних действий...
func (p *Payment) Validate() error {
	var errs []string

	if _, err := NewCurrency(p.Currency.String()); err != nil {
		errs = append(errs, err.Error())
	}

	if p.TotalPrice <= 0 {
		errs = append(errs, "total price must be greater than 0")
	}

	if _, err := NewPaymentMethod(p.PaymentMethod.String()); err != nil {
		errs = append(errs, err.Error())
	}

	if len(p.Description) > MaxPaymentDataLength {
		errs = append(errs, "payment data must be less than 256 characters")
	}

	if _, err := NewStatus(p.Status.String()); err != nil {
		errs = append(errs, err.Error())
	}

	if p.CreatedAt.After(p.UpdatedAt) {
		errs = append(errs, "created at must be before updated at")
	}

	if p.CreatedAt.After(time.Now().UTC()) {
		errs = append(errs, "created at must be before now")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w", fmt.Errorf("%s: %s", ErrInvalidArgument, strings.Join(errs, ", ")))
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

type OrderEvent struct {
	OrderID       string
	UserID        string
	Currency      string
	TotalPrice    float64
	PaymentMethod string
	Description   string
}

func (o *Payment) OrderEvent() OrderEvent {
	return OrderEvent{
		OrderID:       o.OrderID.String(),
		UserID:        o.UserID.String(),
		Currency:      o.Currency.String(),
		TotalPrice:    o.TotalPrice,
		PaymentMethod: o.PaymentMethod.String(),
		Description:   o.Description,
	}
}

func (e OrderEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		OrderID       string `json:"order_id"`
		UserID        string `json:"user_id"`
		Currency      string `json:"currency"`
		TotalPrice    string `json:"total_price"`
		PaymentMethod string `json:"payment_method"`
		Description   string `json:"description"`
	}{
		OrderID:       e.OrderID,
		UserID:        e.UserID,
		Currency:      e.Currency,
		TotalPrice:    fmt.Sprintf("%.2f", e.TotalPrice), // TODO норм ли это
		PaymentMethod: e.PaymentMethod,
		Description:   e.Description,
	})
}

func (e *OrderEvent) UnmarshalJSON(data []byte) error {
	var aux struct {
		OrderID       string  `json:"order_id"`
		UserID        string  `json:"user_id"`
		Currency      string  `json:"currency"`
		TotalPrice    float64 `json:"total_price,string"`
		PaymentMethod string  `json:"payment_method"`
		Description   string  `json:"description"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	e.OrderID = aux.OrderID
	e.UserID = aux.UserID
	e.Currency = aux.Currency
	e.TotalPrice = aux.TotalPrice
	e.PaymentMethod = aux.PaymentMethod
	e.Description = aux.Description
	return nil
}
