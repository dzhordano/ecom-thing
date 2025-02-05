package domain

import (
	"fmt"
	"github.com/google/uuid"
	"math"
	"strings"
	"time"
)

type Product struct {
	ID        uuid.UUID
	Name      string
	Desc      string
	Category  string
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewValidatedProduct(id uuid.UUID, name, description, category string, price float64) (*Product, error) {
	p := NewProduct(id, name, description, category, price)

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

func NewProduct(id uuid.UUID, name, description, category string, price float64) *Product {
	return &Product{
		ID:        id,
		Name:      name,
		Desc:      description,
		Category:  category,
		Price:     price,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

const (
	maxProdCategoryLength = 32
	maxProdNameLength     = 256
	maxProdDescLength     = 2048
)

func (c *Product) Validate() error {
	var errs []string

	if c.Name == "" || len(c.Name) > maxProdNameLength {
		errs = append(errs, "invalid name")
	}

	if c.Desc == "" || len(c.Desc) > maxProdDescLength {
		errs = append(errs, "invalid description")
	}

	if c.Category == "" || len(c.Category) > maxProdCategoryLength {
		errs = append(errs, "invalid category")
	}

	if c.Price <= 0 || math.IsNaN(c.Price) || math.IsInf(c.Price, 0) {
		errs = append(errs, "invalid price")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %s", ErrInvalidArgument, strings.Join(errs, ", "))
	}

	return nil
}

func (c *Product) Update(name, description, category string, price float64) {
	c.Name = name
	c.Desc = description
	c.Category = category
	c.Price = price
	c.UpdatedAt = time.Now()
}
