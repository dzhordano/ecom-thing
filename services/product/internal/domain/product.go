package domain

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxProdCategoryLength = 32
	MaxProdNameLength     = 256
	MaxProdDescLength     = 2048

	MinPrice = 0.01
	MaxPrice = 128000
)

type Product struct {
	ID        uuid.UUID
	Name      string
	Desc      string
	Category  string
	Price     float64
	IsActive  bool
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
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// TODO Вроде неплохо, а вроде и без констант...
func (c *Product) Validate() error {
	var errs []string

	if !ValidateName(c.Name) {
		errs = append(errs, "invalid name")
	}

	if !ValidateDescription(c.Desc) {
		errs = append(errs, "invalid description")
	}

	if !ValidateCategory(c.Category) {
		errs = append(errs, "invalid category")
	}

	if !ValidatePrice(c.Price) {
		errs = append(errs, "invalid price")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

func ValidateName(name string) bool {
	if name == "" || len(name) > MaxProdNameLength {
		return false
	}

	return true
}

func ValidateDescription(desc string) bool {
	if desc == "" || len(desc) > MaxProdDescLength {
		return false
	}

	return true
}

func ValidateCategory(category string) bool {
	if category == "" || len(category) > MaxProdCategoryLength {
		return false
	}

	return true
}

func ValidatePrice(price float64) bool {
	if price < MinPrice || price > MaxPrice || math.IsNaN(price) {
		return false
	}

	return true
}

func (c *Product) Update(name, description, category string, isActive bool, price float64) {
	c.Name = name
	c.Desc = description
	c.Category = category
	c.Price = price
	c.IsActive = isActive
	c.UpdatedAt = time.Now()
}
