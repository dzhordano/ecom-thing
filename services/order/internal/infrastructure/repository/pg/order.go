package pg

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ordersTable     = "orders"
	orderItemsTable = "order_items"
	coupons_talbe   = "coupons"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

// Create implements repository.OrderRepository.
func (o *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
	insertQuery := squirrel.Insert(ordersTable).
		Columns("id", "user_id", "status", "currency", "total_price", "payment_method",
			"delivery_method", "delivery_address", "delivery_date", "items", "created_at", "updated_at").
		Values(order.ID.String(), order.UserID.String(), order.Status, order.Currency, order.TotalPrice, order.PaymentMethod,
			order.DeliveryMethod, order.DeliveryAddress, order.DeliveryDate, order.Items, order.CreatedAt, order.UpdatedAt).
		PlaceholderFormat(squirrel.Dollar)

	query, args, err := insertQuery.ToSql()
	if err != nil {
		return err
	}

	_, err = o.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

// Delete implements repository.OrderRepository.
func (o *OrderRepository) Delete(ctx context.Context, orderId string) error {
	panic("unimplemented")
}

// GetById implements repository.OrderRepository.
func (o *OrderRepository) GetById(ctx context.Context, orderId string) (*domain.Order, error) {
	panic("unimplemented")
}

// GetCoupon implements repository.OrderRepository.
func (o *OrderRepository) GetCoupon(ctx context.Context, code string) (*domain.Coupon, error) {
	panic("unimplemented")
}

// ListByUser implements repository.OrderRepository.
func (o *OrderRepository) ListByUser(ctx context.Context, userId string) ([]*domain.Order, error) {
	panic("unimplemented")
}

// Update implements repository.OrderRepository.
func (o *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
	panic("unimplemented")
}

func NewOrderRepository(db *pgxpool.Pool) repository.OrderRepository {
	return &OrderRepository{
		db: db,
	}
}
