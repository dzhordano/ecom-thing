package service

import (
	"context"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain/repository"
	"github.com/dzhordano/ecom-thing/services/inventory/pkg/logger"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type ItemService struct {
	log  logger.Logger
	repo repository.ItemRepository
}

func NewItemService(log logger.Logger, itemRepository repository.ItemRepository) interfaces.ItemService {
	return &ItemService{
		log:  log,
		repo: itemRepository,
	}
}

func (s *ItemService) GetItem(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	item, err := s.repo.GetItem(ctx, id.String())
	if err != nil {
		s.log.Error("error getting item", "error", err)

		return nil, err
	}

	return item, nil
}

// IsReservable implements interfaces.ItemService.
//
// Function does not return ProductNotFound error due to it's purpose (being called for reservation from order service).
// If product is not found, it returns false.
// Take a note that you have to ensure that ids are NOT duplicated.
func (s *ItemService) IsReservable(ctx context.Context, items map[string]uint64) (bool, error) {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}

	resItems, err := s.repo.GetManyItems(ctx, keys)
	if err != nil {
		s.log.Error("error getting items", "error", err)
		return false, err
	}

	if len(resItems) != len(items) {
		s.log.Debug("not all items found", "got", len(resItems), "expected", len(items))
		return false, nil
	}

	for i := range keys {
		if resItems[i].AvailableQuantity < items[keys[i]] {
			return false, nil
		}
	}

	return true, nil
}

// SetItemWithOp implements interfaces.ItemService.
func (s *ItemService) SetItemWithOp(ctx context.Context, id uuid.UUID, quantity uint64, op string) error {
	item, err := s.repo.GetItem(ctx, id.String())
	if err != nil && op != domain.OperationAdd {
		s.log.Error("error getting item", "error", err)
		return err
	}

	if item == nil {
		item = domain.NewItem(id)
	}

	if err := performOp(item, quantity, op); err != nil {
		s.log.Error("error performing operation", "error", err)
		return errors.Wrap(err, "error performing operation")
	}

	if err := s.repo.SetItem(ctx, item.ProductID.String(), item.AvailableQuantity, item.ReservedQuantity); err != nil {
		s.log.Error("error setting item", "error", err)
		return err
	}

	s.log.Debug("item successfully set", "id", id.String())

	return nil
}

// SetItemsWithOp implements interfaces.ItemService.
func (s *ItemService) SetItemsWithOp(ctx context.Context, items map[string]uint64, op string) error {
	dItems := make([]domain.Item, 0, len(items))
	for id := range items {
		i, err := s.repo.GetItem(ctx, id)
		if err != nil && op != domain.OperationAdd {
			s.log.Error("error getting item", "error", err)

			return err
		}

		if i == nil {
			i = domain.NewItem(uuid.MustParse(id)) // FIXME This may panic. BUT, it actually WONT, unless you delete validUUID method from handlers.
		}

		if err := performOp(i, items[id], op); err != nil {
			s.log.Error("error performing operation", "error", err)

			return err
		}

		dItems = append(dItems, *i)
	}

	if err := s.repo.SetManyItems(ctx, dItems); err != nil {
		s.log.Error("error setting items", "error", err)

		return err
	}

	s.log.Debug("items successfully set", "count", len(items))

	return nil
}

// performOp performs operation on item (i.e. add, sub, lock, unlock, sub_locked)
func performOp(item *domain.Item, quantity uint64, op string) error {
	switch op {
	case domain.OperationAdd:
		item.AddQuantity(quantity)
		return nil
	case domain.OperationSub:
		return item.SubQuantity(quantity)
	case domain.OperationLock:
		return item.LockQuantity(quantity)
	case domain.OperationUnlock:
		return item.UnlockQuantity(quantity)
	case domain.OperationSubLocked:
		return item.SubLockedQuantity(quantity)
	default:
		return domain.ErrOperationUnknown
	}
}
