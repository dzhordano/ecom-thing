package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	MaxQueryLength = 256

	MaxLimit = 100

	DefaultLimit  = 20
	DefaultOffset = 0
)

type SearchParams struct {
	Query            *string
	Description      *string
	Status           *string
	Currency         *string
	MinPrice         *float64
	MaxPrice         *float64
	PaymentMethod    *string
	DeliveryMethod   *string
	DeliveryAddress  *string
	DeliveryDateFrom time.Time
	DeliveryDateTo   time.Time
	MinItemsAmount   *uint64
	MaxItemsAmount   *uint64
	Limit            uint64
	Offset           uint64
}

func NewSearchParams(filters map[string]any) SearchParams {
	s := SearchParams{}

	q := filters["query"].(*string)
	if q != nil {
		s.Query = q
	}

	d := filters["description"].(*string)
	if d != nil {
		s.Description = d
	}

	st := filters["status"].(*string)
	if st != nil {
		s.Status = st
	}

	c := filters["currency"].(*string)
	if c != nil {
		s.Currency = c
	}

	mnp := filters["minPrice"].(*float64)
	if mnp != nil {
		s.MinPrice = mnp
	}

	mxp := filters["maxPrice"].(*float64)
	if mxp != nil {
		s.MaxPrice = mxp
	}

	pm := filters["paymentMethod"].(*string)
	if pm != nil {
		s.PaymentMethod = pm
	}

	dm := filters["deliveryMethod"].(*string)
	if dm != nil {
		s.DeliveryMethod = dm
	}

	da := filters["deliveryAddress"].(*string)
	if da != nil {
		s.DeliveryAddress = da
	}

	dFrom, ok := filters["deliveryDateFrom"].(time.Time)
	if ok && !dFrom.IsZero() {
		s.DeliveryDateFrom = dFrom
	}

	dTo, ok := filters["deliveryDateTo"].(time.Time)
	if ok && !dTo.IsZero() {
		s.DeliveryDateTo = dTo
	}

	mnia := filters["minItemsAmount"].(*uint64)
	if mnia != nil {
		s.MinItemsAmount = mnia
	}

	mxia := filters["maxItemsAmount"].(*uint64)
	if mxia != nil {
		s.MaxItemsAmount = mxia
	}

	l := filters["limit"].(*uint64)
	if l != nil {
		s.Limit = min(*l, MaxLimit)
	} else {
		s.Limit = DefaultLimit
	}

	o := filters["offset"].(*uint64)
	if o != nil {
		s.Offset = *o
	} else {
		s.Offset = DefaultOffset
	}

	// In case you wanna look:
	// log.Printf(
	// 	"search params:\n query=%v\n description%v\n status=%v\n currency=%v\n minPrice=%v\n maxPrice=%v\n deliveryMethod=%v\n paymentMethod=%v\n deliveryAddress=%v\n deliveryDateFrom=%v\n deliveryDateTo=%v\n minItemsAmount=%v\n maxItemsAmount=%v\n limit=%d\n offset=%d\n",
	// 	*s.Query, *s.Description, *s.Status, *s.Currency, *s.MinPrice, *s.MaxPrice, *s.DeliveryMethod, *s.PaymentMethod, *s.DeliveryAddress, s.DeliveryDateFrom, s.DeliveryDateTo, *s.MinItemsAmount, *s.MaxItemsAmount, s.Limit, s.Offset,
	// )

	return s
}

func (o *SearchParams) Validate() error {
	var errs []string

	if o.Query != nil && (*o.Query == "" || len(*o.Query) > MaxQueryLength) {
		errs = append(errs, "invalid query")
	}

	if o.Description != nil && len(*o.Description) > MaxDescriptionLength {
		errs = append(errs, "invalid description")
	}

	if o.Status != nil {
		s, err := NewStatus(*o.Status)
		if err != nil || !s.IsValid() {
			errs = append(errs, "invalid status")
		}
	}

	if o.Currency != nil {
		c, err := NewCurrency(*o.Currency)
		if err != nil || !c.IsValid() {
			errs = append(errs, "invalid currency")
		}
	}

	if o.PaymentMethod != nil {
		p, err := NewPaymentMethod(*o.PaymentMethod)
		if err != nil || !p.IsValid() {
			errs = append(errs, "invalid payment method")
		}
	}

	if o.DeliveryMethod != nil {
		d, err := NewDeliveryMethod(*o.DeliveryMethod)
		if err != nil || !d.IsValid() {
			errs = append(errs, "invalid delivery method")
		}
	}

	// TODO валидировать адрес при создании соответствующего метода.
	if o.DeliveryAddress != nil && len(*o.DeliveryAddress) < MinAddressLength {
		errs = append(errs, "invalid delivery address")
	}

	if !o.DeliveryDateTo.IsZero() {
		if !o.DeliveryDateFrom.IsZero() && o.DeliveryDateFrom.After(o.DeliveryDateTo) {
			errs = append(errs, "invalid delivery date range")
		}
	}

	if o.MinItemsAmount != nil && *o.MinItemsAmount > *o.MaxItemsAmount {
		errs = append(errs, "invalid items amount range")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalidArgument, strings.Join(errs, ", "))
	}

	return nil
}
