package domain

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"
)

type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       float64
	Quantity    uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewValidatedProduct(id uuid.UUID, name string, description string, price float64, quantity uint64) (*Product, error) {
	p := NewProduct(id, name, description, price, quantity)

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

func NewProduct(id uuid.UUID, name string, description string, price float64, quantity uint64) *Product {
	return &Product{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
		Quantity:    quantity,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (p *Product) Validate() error {
	var errs []string

	if p.Name == "" || len(p.Name) > 256 {
		errs = append(errs, "invalid name")
	}

	if p.Description == "" || len(p.Description) > 2048 {
		errs = append(errs, "invalid description")
	}

	if p.Price <= 0 {
		errs = append(errs, "invalid price")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalidArgument, strings.Join(errs, ", "))
	}

	return nil
}
