package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxDescriptionLength = 255

	MinAddressLength = 5
)

type Order struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Description     string
	Status          Status
	Currency        Currency
	TotalPrice      float64
	PaymentMethod   PaymentMethod
	DeliveryMethod  DeliveryMethod
	DeliveryAddress string
	DeliveryDate    time.Time
	Items           Items
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewOrder(userId uuid.UUID, description, status, currency string, totalPrice, totalDiscount float64,
	paymentMethod, deliveryMethod, deliveryAddress string, deliveryDate time.Time,
	items Items) (*Order, error) {

	orderId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	o := &Order{
		ID:              orderId,
		UserID:          userId,
		Description:     description,
		Status:          Status(status),
		Currency:        Currency(currency),
		TotalPrice:      ApplyDiscountTo(totalPrice, totalDiscount),
		PaymentMethod:   PaymentMethod(paymentMethod),
		DeliveryMethod:  DeliveryMethod(deliveryMethod),
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
	var errs []string

	if len(o.Description) > MaxDescriptionLength {
		errs = append(errs, ErrInvalidDescription.Error())
	}

	if o.TotalPrice < 0 {
		errs = append(errs, ErrInvalidPrice.Error())
	}

	if !o.Status.IsValid() {
		errs = append(errs, ErrInvalidOrderStatus.Error())
	}

	if !o.Currency.IsValid() {
		errs = append(errs, ErrInvalidCurrency.Error())
	}

	if !o.PaymentMethod.IsValid() {
		errs = append(errs, ErrInvalidPaymentMethod.Error())
	}

	if !o.DeliveryMethod.IsValid() {
		errs = append(errs, ErrInvalidDeliveryMethod.Error())
	}

	// TODO Нужен REGEX или что-то такое. Мб вообще проверять на существование такой улицы...
	if len(o.DeliveryAddress) < MinAddressLength {
		errs = append(errs, ErrInvalidDeliveryAddress.Error())
	}

	if o.DeliveryDate.Before(time.Now()) {
		errs = append(errs, ErrInvalidDeliveryDate.Error())
	}

	if len(o.Items) == 0 {
		errs = append(errs, ErrInvalidOrderItems.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w", fmt.Errorf("%s: %s", ErrInvalidArgument, strings.Join(errs, ", ")))
	}

	return nil
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

func (i Item) Value() (driver.Value, error) {
	return fmt.Sprintf("\"(%s,%d)\"", i.ProductID, i.Quantity), nil
}

type Items []Item

func (items Items) Value() (driver.Value, error) {
	parts := make([]string, len(items))
	for i, item := range items {
		v, err := item.Value()
		if err != nil {
			return nil, err
		}
		parts[i] = v.(string)
	}

	s := strings.Join(parts, ",")

	return fmt.Sprintf("{%s}", s), nil
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
	BankCard PaymentMethod = "bank_card"
	Cash     PaymentMethod = "cash"
)

var validPaymentMethods = map[PaymentMethod]bool{
	BankCard: true,
	Cash:     true,
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

type Coupon struct {
	ID        uint
	Code      string
	Discount  float64
	ValidFrom time.Time
	ValidTo   time.Time
}

type OrderEvent struct {
	OrderID       string
	UserID        string
	Currency      string
	TotalPrice    string
	PaymentMethod string
	Description   string
}

type InventoryEvent struct {
	OrderID string
	Items   Items
}

func (o *Order) OrderEvent() OrderEvent {
	return OrderEvent{
		OrderID:       o.ID.String(),
		UserID:        o.UserID.String(),
		Currency:      o.Currency.String(),
		TotalPrice:    fmt.Sprintf("%.2f", o.TotalPrice),
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
		TotalPrice:    e.TotalPrice,
		PaymentMethod: e.PaymentMethod,
		Description:   e.Description,
	})
}

func (o *Order) InventoryEvent() InventoryEvent {
	return InventoryEvent{
		OrderID: o.ID.String(),
		Items:   o.Items,
	}
}

func (e InventoryEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		OrderID string `json:"order_id"`
		Items   Items  `json:"items"`
	}{
		OrderID: e.OrderID,
		Items:   e.Items,
	})
}
