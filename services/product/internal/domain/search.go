package domain

import (
	"fmt"
	"strings"
)

const (
	MaxQueryLength = 256

	MaxLimit = 100

	DefaultLimit  = 20
	DefaultOffset = 0
)

type SearchParams struct {
	Query    *string
	Category *string
	MinPrice *float64
	MaxPrice *float64
	Limit    uint64
	Offset   uint64
}

func NewSearchParams(filters map[string]any) SearchParams {
	s := SearchParams{}

	q, ok := filters["query"].(*string)
	if ok {
		s.Query = q
	}

	c, ok := filters["category"].(*string)
	if ok {
		s.Category = c
	}

	mn, ok := filters["minPrice"].(*float64)
	if ok {
		s.MinPrice = mn
	}

	mx, ok := filters["maxPrice"].(*float64)
	if ok {
		s.MaxPrice = mx
	}

	l, ok := filters["limit"].(*uint64)
	if !ok {
		s.Limit = DefaultLimit
	} else if l != nil {
		s.Limit = min(*l, MaxLimit)
	}

	o, ok := filters["offset"].(*uint64)
	if !ok {
		s.Offset = DefaultOffset
	} else if o != nil {
		s.Offset = *o
	}

	return s
}

func (o *SearchParams) Validate() error {
	var errs []string

	if o.Query != nil && (*o.Query == "" || len(*o.Query) > MaxQueryLength) {
		errs = append(errs, "invalid query")
	}

	if o.Category != nil && !ValidateCategory(*o.Category) {
		errs = append(errs, "invalid category")
	}

	if o.MinPrice != nil && !ValidatePrice(*o.MinPrice) {
		errs = append(errs, "invalid min price")
	}

	if o.MaxPrice != nil && !ValidatePrice(*o.MaxPrice) {
		errs = append(errs, "invalid max price")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalidArgument, strings.Join(errs, ", "))
	}

	return nil
}
