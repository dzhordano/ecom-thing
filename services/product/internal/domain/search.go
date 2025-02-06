package domain

import (
	"fmt"
	"strings"
)

type SearchOptions struct {
	Query    *string
	Category *string
	MinPrice *float64
	MaxPrice *float64
}

func NewSearchOptions(query *string, category *string, minPrice *float64, maxPrice *float64) SearchOptions {
	return SearchOptions{
		Query:    query,
		Category: category,
		MinPrice: minPrice,
		MaxPrice: maxPrice,
	}
}

func (o *SearchOptions) Validate() error {
	var errs []string

	if o.Query != nil && *o.Query == "" {
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
