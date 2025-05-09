package service

import (
	"context"
	"errors"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain/repository"
	api "github.com/dzhordano/ecom-thing/services/inventory/pkg/api/inventory/v1"
	"github.com/dzhordano/ecom-thing/services/inventory/pkg/logger"
	"github.com/google/uuid"
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
		s.log.Error("error getting item", "error", err, "item_id", id.String())
		return nil, domain.NewAppError(err, "failed to get item")
	}

	return item, nil
}

// IsReservable implements interfaces.ItemService.
//
// Function does not return ProductNotFound error due to it's purpose (being called for reservation from order service).
// If product is not found, it returns false.
// Note that you have to ensure that the ids are NOT duplicated.
func (s *ItemService) IsReservable(ctx context.Context, items map[string]uint64) (bool, error) {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}

	resItems, err := s.repo.GetManyItems(ctx, keys)
	if err != nil {
		s.log.Error("error getting items", "error", err)
		return false, domain.NewAppError(err, "failed to get items")
	}

	if len(resItems) != len(items) {
		s.log.Debug("not all items found", "got", len(resItems), "expected", len(items))
		return false, domain.NewAppError(domain.ErrProductNotFound, "not all items found")
	}

	for i := range keys {
		if resItems[i].AvailableQuantity < items[keys[i]] {
			s.log.Debug("not enough quantity", "got", resItems[i].AvailableQuantity, "need", items[keys[i]])
			return false, nil
		}
	}

	s.log.Debug("all items found", "got", len(resItems), "expected", len(items))

	return true, nil
}

// SetItemWithOp implements interfaces.ItemService.
func (s *ItemService) SetItemWithOp(ctx context.Context, id uuid.UUID, quantity uint64, op string) error {

	dOp := protoEnumToDomainOp(op)

	item, err := s.repo.GetItem(ctx, id.String())
	if err != nil && dOp != domain.OperationAdd {
		s.log.Error("error getting item", "error", err, "item_id", id.String())
		return domain.NewAppError(err, "failed to get item")
	}

	if item == nil {
		item = domain.NewItem(id)
	}

	if err := performOp(item, quantity, dOp); err != nil {
		s.log.Error("error performing operation", "error", err, "item_id", id.String())
		return domain.NewAppError(err, err.Error())
	}

	if err := s.repo.SetItem(ctx, item.ProductID.String(), item.AvailableQuantity, item.ReservedQuantity); err != nil {
		s.log.Error("error setting item", "error", err, "item_id", id.String())
		return domain.NewAppError(err, "failed to set item")
	}

	s.log.Debug("item successfully set", "id", id.String())

	return nil
}

// SetItemsWithOp implements interfaces.ItemService.
func (s *ItemService) SetItemsWithOp(ctx context.Context, items map[string]uint64, op string) error {
	dItems := make([]domain.Item, 0, len(items))
	dOp := protoEnumToDomainOp(op)

	// Flag for checking if operation is 'add'.
	isOpAdd := dOp == domain.OperationAdd

	// TODO optimize?
	for id := range items {
		i, err := s.repo.GetItem(ctx, id)
		if err != nil {
			// If a product is not found and operation is added, it's ok.
			if errors.Is(err, domain.ErrProductNotFound) && !isOpAdd {
				s.log.Error("error getting item", "error", err, "item_id", id)
				return domain.NewAppError(err, "failed to get item")
			}
		}

		if i == nil {
			i = domain.NewItem(uuid.MustParse(id)) // WARNING: This may panic. BUT, it actually WONT unless you delete validUUID method from handlers.
		}

		if err := performOp(i, items[id], dOp); err != nil {
			s.log.Error("error performing operation", "error", err, "item_id", id)

			return domain.NewAppError(err, err.Error())
		}

		dItems = append(dItems, *i)
	}

	if err := s.repo.SetManyItems(ctx, dItems); err != nil {
		s.log.Error("error setting items", "error", err)

		return domain.NewAppError(err, "failed to set items")
	}

	s.log.Debug("items successfully set", "count", len(items), "op", op)

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

func protoEnumToDomainOp(op string) string {
	switch op {
	case api.OperationType_OPERATION_TYPE_ADD.String():
		return domain.OperationAdd
	case api.OperationType_OPERATION_TYPE_SUB.String():
		return domain.OperationSub
	case api.OperationType_OPERATION_TYPE_LOCK.String():
		return domain.OperationLock
	case api.OperationType_OPERATION_TYPE_UNLOCK.String():
		return domain.OperationUnlock
	case api.OperationType_OPERATION_TYPE_SUB_LOCKED.String():
		return domain.OperationSubLocked
	default:
		return domain.OperationUnknown
	}
}
