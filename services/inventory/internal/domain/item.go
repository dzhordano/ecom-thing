package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrNotEnoughQuantity = errors.New("not enough quantity")
)

type Item struct {
	ProductID         uuid.UUID
	AvailableQuantity uint64
	ReservedQuantity  uint64
}

func NewItem(productID uuid.UUID, availableQuantity uint64) *Item {
	return &Item{
		ProductID:         productID,
		AvailableQuantity: availableQuantity,
	}
}

func (i *Item) LockQuantity(quantity uint64) error {
	if i.AvailableQuantity < quantity {
		return ErrNotEnoughQuantity
	}
	i.AvailableQuantity -= quantity
	i.ReservedQuantity += quantity
	return nil
}

func (i *Item) UnlockQuantity(quantity uint64) error {
	if i.ReservedQuantity < quantity {
		return ErrNotEnoughQuantity
	}
	i.AvailableQuantity += quantity
	i.ReservedQuantity -= quantity
	return nil
}

func (i *Item) AddQuantity(quantity uint64) {
	i.AvailableQuantity += quantity
}

func (i *Item) SubQuantity(quantity uint64) error {
	if i.AvailableQuantity < quantity {
		return ErrNotEnoughQuantity
	}
	i.AvailableQuantity -= quantity
	return nil
}

func (i *Item) SubLockedQuantity(quantity uint64) error {
	if i.ReservedQuantity < quantity {
		return ErrNotEnoughQuantity
	}
	i.ReservedQuantity -= quantity
	return nil
}
