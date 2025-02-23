package domain

import (
	"fmt"
	"strings"
)

const (
	MaxQueryLength = 256
)

type SearchOptions struct {
	Query    *string
	Category *string
	MinPrice *float64
	MaxPrice *float64
}

func NewSearchOptions(query *string, category *string, minPrice *float64, maxPrice *float64) SearchOptions {
	s := SearchOptions{}

	if query != nil {
		s.Query = query
	}
	if category != nil {
		s.Category = category
	}
	if minPrice != nil {
		s.MinPrice = minPrice
	}
	if maxPrice != nil {
		s.MaxPrice = maxPrice
	}

	return s
}

func (o *SearchOptions) Validate() error {
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
