package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Ошибки для валидации.
var (
	ErrInvalidOrderStatus    = errors.New("invalid order status")
	ErrInvalidCurrency       = errors.New("invalid currency")
	ErrInvalidPaymentMethod  = errors.New("invalid payment method")
	ErrInvalidDeliveryMethod = errors.New("invalid delivery method")

	ErrInternal               = errors.New("internal error")
	ErrInvalidUUID            = errors.New("invalid uuid")
	ErrInvalidPrice           = errors.New("invalid price")
	ErrInvalidDiscount        = errors.New("invalid discount")
	ErrInvalidDeliveryAddress = errors.New("invalid delivery address")
	ErrInvalidDeliveryDate    = errors.New("invalid delivery date")
	ErrInvalidOrderItems      = errors.New("invalid order items")
)

type Order struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Status          Status
	Currency        Currency
	TotalPrice      float64
	PaymentMethod   PaymentMethod
	DeliveryMethod  DeliveryMethod
	DeliveryAddress string
	DeliveryDate    time.Time
	Items           []Item
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewOrder(userId uuid.UUID, status, currency string, totalPrice, totalDiscount float64,
	paymentMethod, deliveryMethod, deliveryAddress string, deliveryDate time.Time,
	items []Item) (*Order, error) {

	s, err := NewStatus(status)
	if err != nil {
		return nil, err
	}

	c, err := NewCurrency(currency)
	if err != nil {
		return nil, err
	}

	pm, err := NewPaymentMethod(paymentMethod)
	if err != nil {
		return nil, err
	}

	dm, err := NewDeliveryMethod(deliveryMethod)
	if err != nil {
		return nil, err
	}

	orderId, err := uuid.NewUUID()
	if err != nil {
		// FIXME Логировать ТУТ ЖБ.
		return nil, ErrInternal
	}

	if totalDiscount < 0 || totalDiscount > 100 {
		return nil, ErrInvalidDiscount
	}

	o := &Order{
		ID:              orderId,
		UserID:          userId,
		Status:          s,
		Currency:        c,
		TotalPrice:      ApplyDiscountTo(totalPrice, totalDiscount),
		PaymentMethod:   pm,
		DeliveryMethod:  dm,
		DeliveryAddress: deliveryAddress,
		DeliveryDate:    deliveryDate,
		Items:           items,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := o.Validate(); err != nil {
		return nil, err
	}

	return o, nil
}

func (o *Order) Validate() error {
	var err error

	if o.TotalPrice < 0 {
		err = errors.Join(err, ErrInvalidPrice)
	}

	if len(o.DeliveryAddress) < 1 {
		err = errors.Join(err, ErrInvalidDeliveryAddress)
	}

	if o.DeliveryDate.Before(time.Now()) {
		err = errors.Join(err, ErrInvalidDeliveryDate)
	}

	if len(o.Items) == 0 {
		err = errors.Join(err, ErrInvalidOrderItems)
	}

	return err
}

type Status string

const (
	OrderPending   Status = "pending"   // Order was created but not paid.
	OrderPaid      Status = "paid"      // Order is paid but not yet delivered.
	OrderCompleted Status = "completed" // Order was delivered and marked as completed.
	OrderCancelled Status = "cancelled" // Order was cancelled either by user or by some reason.
)

var validStatuses = map[Status]bool{
	OrderPending:   true,
	OrderPaid:      true,
	OrderCompleted: true,
	OrderCancelled: true,
}

// NewStatus создает статус с валидацией.
func NewStatus(status string) (Status, error) {
	s := Status(status)
	if !s.IsValid() {
		return "", ErrInvalidOrderStatus
	}
	return s, nil
}

func (s Status) IsValid() bool {
	_, exists := validStatuses[s]
	return exists
}

func (s Status) String() string {
	return string(s)
}

type Item struct {
	ProductID uuid.UUID
	Quantity  uint64
}

type Currency string

const (
	RUB Currency = "RUB"
	EUR Currency = "EUR"
	USD Currency = "USD"
)

var validCurrencies = map[Currency]bool{
	RUB: true,
	EUR: true,
	USD: true,
}

func NewCurrency(c string) (Currency, error) {
	curr := Currency(c)
	if !curr.IsValid() {
		return "", ErrInvalidCurrency
	}
	return curr, nil
}

func (c Currency) IsValid() bool {
	_, exists := validCurrencies[c]
	return exists
}

func (c Currency) String() string {
	return string(c)
}

type PaymentMethod string

const (
	Card   PaymentMethod = "card"
	PayPal PaymentMethod = "paypal"
	Crypto PaymentMethod = "crypto"
)

var validPaymentMethods = map[PaymentMethod]bool{
	Card:   true,
	PayPal: true,
	Crypto: true,
}

func NewPaymentMethod(p string) (PaymentMethod, error) {
	method := PaymentMethod(p)
	if !method.IsValid() {
		return "", ErrInvalidPaymentMethod
	}
	return method, nil
}

func (p PaymentMethod) IsValid() bool {
	_, exists := validPaymentMethods[p]
	return exists
}

func (p PaymentMethod) String() string {
	return string(p)
}

type DeliveryMethod string

const (
	Standard DeliveryMethod = "standard"
	Express  DeliveryMethod = "express"
	Pickup   DeliveryMethod = "pickup"
)

var validDeliveryMethods = map[DeliveryMethod]bool{
	Standard: true,
	Express:  true,
	Pickup:   true,
}

func NewDeliveryMethod(d string) (DeliveryMethod, error) {
	method := DeliveryMethod(d)
	if !method.IsValid() {
		return "", ErrInvalidDeliveryMethod
	}
	return method, nil
}

func (d DeliveryMethod) IsValid() bool {
	_, exists := validDeliveryMethods[d]
	return exists
}

func (d DeliveryMethod) String() string {
	return string(d)
}

func ApplyDiscountTo(price, discount float64) float64 {
	return price - (price * discount / 100)
}
