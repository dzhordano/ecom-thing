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
		Columns("id", "user_id", "status", "currency", "total_price", "payment_method",
			"delivery_method", "delivery_address", "delivery_date", "items", "created_at", "updated_at").
		Values(order.ID.String(), order.UserID.String(), order.Status, order.Currency, order.TotalPrice, order.PaymentMethod,
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

	selectQuery := sq.Select("id", "user_id", "status", "currency", "total_price", "payment_method",
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
	if err := o.db.QueryRow(ctx, query, args...).Scan(&order.ID, &order.UserID, &order.Status, &order.Currency, &order.TotalPrice, &order.PaymentMethod,
		&order.DeliveryMethod, &order.DeliveryAddress, &order.DeliveryDate, &items, &order.CreatedAt, &order.UpdatedAt); err != nil {
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

	selectQuery := sq.Select("id", "user_id", "status", "currency", "total_price", "payment_method",
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
		if err := rows.Scan(&order.ID, &order.UserID, &order.Status, &order.Currency, &order.TotalPrice, &order.PaymentMethod,
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
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetCoupon implements repository.OrderRepository.
func (o *OrderRepository) GetCoupon(ctx context.Context, code string) (*domain.Coupon, error) {
	selectQuery := sq.Select("id", "code", "discount", "valid_from", "valid_to").
		From(couponsTable).
		Where(sq.Eq{"code": code}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := selectQuery.ToSql()
	if err != nil {
		return nil, err
	}

	var disc domain.Coupon
	if err := o.db.QueryRow(ctx, query, args...).Scan(&disc.ID, &disc.Code, &disc.Discount, &disc.ValidFrom, &disc.ValidTo); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCouponNotFound
		}
		return nil, err
	}

	return &disc, nil
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

func uint64FromString(s string) uint64 {
	i, _ := strconv.ParseUint(s, 10, 64)
	return i
}
