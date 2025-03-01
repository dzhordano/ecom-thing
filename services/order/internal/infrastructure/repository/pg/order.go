package pg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ordersTable  = "orders"
	couponsTable = "coupons"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) repository.OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

// Save implements repository.OrderRepository.
func (o *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
	const op = "repository.OrderRepository.Create"

	insertQuery := sq.Insert(ordersTable).
		Columns("id", "user_id", "description", "status", "currency", "total_price", "payment_method",
			"delivery_method", "delivery_address", "delivery_date", "items", "created_at", "updated_at").
		Values(order.ID.String(), order.UserID.String(), order.Description, order.Status, order.Currency, order.TotalPrice, order.PaymentMethod,
			order.DeliveryMethod, order.DeliveryAddress, order.DeliveryDate, order.Items, order.CreatedAt, order.UpdatedAt).
		PlaceholderFormat(sq.Dollar)

	query, args, err := insertQuery.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = o.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetById implements repository.OrderRepository.
func (o *OrderRepository) GetById(ctx context.Context, orderId string) (*domain.Order, error) {
	const op = "repository.OrderRepository.GetById"

	selectQuery := sq.Select("id", "user_id", "description", "status", "currency", "total_price", "payment_method",
		"delivery_method", "delivery_address", "delivery_date", "items", "created_at", "updated_at").
		From(ordersTable).
		Where(sq.Eq{"id": orderId}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var order domain.Order
	var items []string
	if err := o.db.QueryRow(ctx, query, args...).Scan(
		&order.ID,
		&order.UserID,
		&order.Description,
		&order.Status,
		&order.Currency,
		&order.TotalPrice,
		&order.PaymentMethod,
		&order.DeliveryMethod,
		&order.DeliveryAddress,
		&order.DeliveryDate,
		&items,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrOrderNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	parseItems(&order, items)

	return &order, nil
}

// ListByUser implements repository.OrderRepository.
func (o *OrderRepository) ListByUser(ctx context.Context, userId string) ([]*domain.Order, error) {
	const op = "repository.OrderRepository.ListByUser"

	selectQuery := sq.Select("id", "user_id", "description", "status", "currency", "total_price", "payment_method",
		"delivery_method", "delivery_address", "delivery_date", "items", "created_at", "updated_at").
		From(ordersTable).
		Where(sq.Eq{"user_id": userId}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := o.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var orders []*domain.Order
	for rows.Next() {
		var order domain.Order
		var items []string
		if err := rows.Scan(&order.ID, &order.UserID, &order.Description, &order.Status, &order.Currency, &order.TotalPrice, &order.PaymentMethod,
			&order.DeliveryMethod, &order.DeliveryAddress, &order.DeliveryDate, &items, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		parseItems(&order, items)

		orders = append(orders, &order)
	}

	return orders, nil
}

// TODO Хорошо ли? Узнать про bloat.
//
// Update implements repository.OrderRepository.
func (o *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	const op = "repository.OrderRepository.Update"

	updateQuery := sq.Update(ordersTable).
		Set("description", order.Description).
		Set("status", order.Status).
		Set("currency", order.Currency).
		Set("total_price", order.TotalPrice).
		Set("payment_method", order.PaymentMethod).
		Set("delivery_method", order.DeliveryMethod).
		Set("delivery_address", order.DeliveryAddress).
		Set("delivery_date", order.DeliveryDate).
		Set("items", order.Items).
		Set("updated_at", order.UpdatedAt).
		Where(sq.Eq{"id": order.ID}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := updateQuery.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = o.db.Exec(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, domain.ErrOrderNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// TODO Тут пока просто удаляем.
// Delete implements repository.OrderRepository.
func (o *OrderRepository) Delete(ctx context.Context, orderId string) error {
	const op = "repository.OrderRepository.Delete"

	deleteQuery := sq.Delete(ordersTable).
		Where(sq.Eq{"id": orderId}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := deleteQuery.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = o.db.Exec(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, domain.ErrOrderNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetCoupon implements repository.OrderRepository.
func (o *OrderRepository) GetCoupon(ctx context.Context, code string) (*domain.Coupon, error) {
	const op = "repository.OrderRepository.GetCoupon"

	selectQuery := sq.Select("id", "code", "discount", "valid_from", "valid_to").
		From(couponsTable).
		Where(sq.Eq{"code": code}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var disc domain.Coupon
	if err := o.db.QueryRow(ctx, query, args...).Scan(&disc.ID, &disc.Code, &disc.Discount, &disc.ValidFrom, &disc.ValidTo); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %v: %w", op, err, domain.ErrCouponNotFound) // FIXME Думаю вот так больше контекста??? А то терять саму ошибку не очень...
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &disc, nil
}

func (o *OrderRepository) Search(ctx context.Context, params domain.SearchParams) ([]*domain.Order, error) {
	const op = "repository.OrderRepository.Search"

	selectQuery := sq.Select("id", "user_id", "description", "status", "currency", "total_price", "payment_method", "delivery_method",
		"delivery_address", "delivery_date", "items", "created_at", "updated_at").
		From(ordersTable).
		PlaceholderFormat(sq.Dollar).
		Limit(params.Limit).
		Offset(params.Offset)

	// FIXME проверить производительность запроса.
	// Узнать как такие методы вообще делать/нужны ли они.
	if params.Query != nil {
		selectQuery = selectQuery.Where(sq.Or{
			sq.Like{"description": fmt.Sprintf("%%%s%%", *params.Query)},
			sq.Like{"status": fmt.Sprintf("%%%s%%", *params.Query)},
			sq.Like{"currency": fmt.Sprintf("%%%s%%", *params.Query)},
			sq.Like{"payment_method": fmt.Sprintf("%%%s%%", *params.Query)},
			sq.Like{"delivery_method": fmt.Sprintf("%%%s%%", *params.Query)},
			sq.Like{"delivery_address": fmt.Sprintf("%%%s%%", *params.Query)},
		})
	}

	// TODO оптимизировать эту шляпу?

	if params.Description != nil {
		selectQuery = selectQuery.Where(sq.Eq{"description": *params.Description})
	}

	if params.Status != nil {
		selectQuery = selectQuery.Where(sq.Eq{"status": *params.Status})
	}

	if params.Currency != nil {
		selectQuery = selectQuery.Where(sq.Eq{"currency": *params.Currency})
	}

	if params.PaymentMethod != nil {
		selectQuery = selectQuery.Where(sq.Eq{"payment_method": *params.PaymentMethod})
	}

	if params.DeliveryMethod != nil {
		selectQuery = selectQuery.Where(sq.Eq{"delivery_method": *params.DeliveryMethod})
	}

	if params.DeliveryAddress != nil {
		selectQuery = selectQuery.Where(sq.Eq{"delivery_address": *params.DeliveryAddress})
	}

	if !params.DeliveryDateFrom.IsZero() {
		selectQuery = selectQuery.Where(sq.GtOrEq{"delivery_date": params.DeliveryDateFrom})
	}

	if !params.DeliveryDateTo.IsZero() {
		selectQuery = selectQuery.Where(sq.LtOrEq{"delivery_date": params.DeliveryDateTo})
	}

	if params.MinPrice != nil {
		selectQuery = selectQuery.Where(sq.GtOrEq{"total_price": *params.MinPrice})
	}

	if params.MaxPrice != nil {
		selectQuery = selectQuery.Where(sq.LtOrEq{"total_price": *params.MaxPrice})
	}

	if params.MinItemsAmount != nil {
		selectQuery = selectQuery.Where(sq.GtOrEq{"(array_length(items, 1))": *params.MinItemsAmount})
	}

	if params.MaxItemsAmount != nil {
		selectQuery = selectQuery.Where(sq.LtOrEq{"(array_length(items, 1))": *params.MaxItemsAmount})
	}

	query, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := o.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var orders []*domain.Order
	for rows.Next() {
		var order domain.Order
		var items []string
		if err := rows.Scan(&order.ID, &order.UserID, &order.Description, &order.Status, &order.Currency, &order.TotalPrice, &order.PaymentMethod, &order.DeliveryMethod,
			&order.DeliveryAddress, &order.DeliveryDate, &items, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		parseItems(&order, items)

		orders = append(orders, &order)
	}

	return orders, nil
}

// parseItems преобразует строку полученную из бд в []domain.Item.
func parseItems(order *domain.Order, items []string) {
	for _, item := range items {
		sitems := strings.Split(item, ",")

		pId := (sitems[0])[1:]
		pQuantity := (sitems[1])[:(len(sitems[1]) - 1)]

		order.Items = append(order.Items, domain.Item{
			ProductID: uuid.MustParse(pId), // TODO По идее невозможна паника, так как БД консистентна...?
			Quantity:  uint64FromString(pQuantity),
		})
	}
}

// Эта функция подразумевает уже "правильный" ввод, так как поле заведомо валидно.
//
// НЕ использовать если это не так.
func uint64FromString(s string) uint64 {
	i, _ := strconv.ParseUint(s, 10, 64)
	return i
}
