package service

import (
	"context"
	"errors"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain/repository"
	"github.com/dzhordano/ecom-thing/services/inventory/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ItemService struct {
	log  logger.BaseLogger
	repo repository.ItemRepository
}

func NewItemService(log logger.BaseLogger, itemRepository repository.ItemRepository) interfaces.ItemService {
	return &ItemService{
		log:  log,
		repo: itemRepository,
	}
}

func (s *ItemService) GetItem(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	item, err := s.repo.GetItem(ctx, id.String())
	if err != nil {
		s.log.Error("error getting item", zap.Error(err))
		return nil, err
	}

	return item, nil
}

func (s *ItemService) AddItemQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error {
	item, err := s.repo.GetItem(ctx, id.String())
	if err != nil {
		if !errors.Is(err, domain.ErrProductNotFound) {
			s.log.Error("error getting item", zap.Error(err))
			return err
		}
	}

	if nil == item {
		item = domain.NewItem(id, quantity)
	} else {
		item.AddQuantity(quantity)
	}

	if err = s.repo.SetItem(ctx, id.String(), item.AvailableQuantity, item.ReservedQuantity); err != nil {
		s.log.Error("error setting item", zap.Error(err))
		return err
	}

	return nil
}

func (s *ItemService) SubItemQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error {
	item, err := s.repo.GetItem(ctx, id.String())
	if err != nil {
		s.log.Error("error getting item", zap.Error(err))
		return err
	}

	if err = item.SubQuantity(quantity); err != nil {
		s.log.Error("error subtracting item quantity", zap.Error(err))
		return err
	}

	if err = s.repo.SetItem(ctx, id.String(), item.AvailableQuantity, item.ReservedQuantity); err != nil {
		s.log.Error("error setting item", zap.Error(err))
		return err
	}

	return nil
}

func (s *ItemService) LockItemQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error {
	item, err := s.repo.GetItem(ctx, id.String())
	if err != nil {
		s.log.Error("error getting item", zap.Error(err))
		return err
	}

	if err = item.LockQuantity(quantity); err != nil {
		s.log.Error("error locking item quantity", zap.Error(err))
		return err
	}

	if err = s.repo.SetItem(ctx, id.String(), item.AvailableQuantity, item.ReservedQuantity); err != nil {
		s.log.Error("error setting item", zap.Error(err))
		return err
	}

	return nil
}

func (s *ItemService) UnlockItemQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error {
	item, err := s.repo.GetItem(ctx, id.String())
	if err != nil {
		s.log.Error("error getting item", zap.Error(err))
		return err
	}

	if err = item.UnlockQuantity(quantity); err != nil {
		s.log.Error("error unlocking item quantity", zap.Error(err))
		return err
	}

	if err = s.repo.SetItem(ctx, id.String(), item.AvailableQuantity, item.ReservedQuantity); err != nil {
		s.log.Error("error setting item", zap.Error(err))
		return err
	}

	return nil
}

func (s *ItemService) SubLockedItemQuantity(ctx context.Context, id uuid.UUID, quantity uint64) error {
	item, err := s.repo.GetItem(ctx, id.String())
	if err != nil {
		s.log.Error("error getting item", zap.Error(err))
		return err
	}

	if err = item.SubLockedQuantity(quantity); err != nil {
		s.log.Error("error subtracting locker item quantity", zap.Error(err))
		return err
	}

	if err = s.repo.SetItem(ctx, id.String(), item.AvailableQuantity, item.ReservedQuantity); err != nil {
		s.log.Error("error setting item", zap.Error(err))
		return err
	}

	return nil
}
