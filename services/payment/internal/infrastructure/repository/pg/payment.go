package pg

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	paymentsTable         = "payments"
	paymentsStatusesTable = "payments_statuses"
	outboxTable           = "outbox"
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) repository.PaymentRepository {
	return &PaymentRepository{
		db: db,
	}
}

// Save implements repository.PaymentRepository.
// FIXME outbox + tx
func (r *PaymentRepository) Save(ctx context.Context, payment *domain.Payment) error {
	insQuery := sq.Insert(paymentsTable).
		Columns("id", "user_id", "order_id", "currency", "total_price", "status", "payment_method", "description", "created_at", "updated_at").
		Values(payment.ID, payment.UserID, payment.OrderID, payment.Currency, payment.TotalPrice, payment.Status, payment.PaymentMethod, payment.Description, payment.CreatedAt, payment.UpdatedAt).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	query, args, err := insQuery.ToSql()
	if err != nil {
		return err
	}

	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

// GetById implements repository.PaymentRepository.
func (r *PaymentRepository) GetById(ctx context.Context, paymentId string, userId string) (*domain.Payment, error) {
	selQuery := sq.Select("id", "user_id", "order_id", "currency", "total_price", "status", "payment_method", "description", "status", "created_at", "updated_at").
		From(paymentsTable).
		Where(sq.Eq{"id": paymentId, "user_id": userId}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := selQuery.ToSql()
	if err != nil {
		return nil, err
	}

	var payment domain.Payment
	if err = r.db.QueryRow(ctx, query, args...).Scan(
		&payment.ID,
		&payment.UserID,
		&payment.OrderID,
		&payment.Currency,
		&payment.TotalPrice,
		&payment.Status,
		&payment.PaymentMethod,
		&payment.Description,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &payment, nil
}

// ListByUser implements repository.PaymentRepository.
func (r *PaymentRepository) ListByUser(ctx context.Context, userId string, limit uint64, offset uint64) ([]*domain.Payment, error) {
	selQuery := sq.Select("id", "user_id", "order_id", "currency", "total_price", "status", "payment_method", "description", "status", "created_at", "updated_at").
		From(paymentsTable).
		Where(sq.Eq{"user_id": userId}).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar)

	query, args, err := selQuery.ToSql()
	if err != nil {
		return nil, err
	}

	var payments []*domain.Payment
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var payment domain.Payment
		if err := rows.Scan(
			&payment.ID,
			&payment.UserID,
			&payment.OrderID,
			&payment.Currency,
			&payment.TotalPrice,
			&payment.Status,
			&payment.PaymentMethod,
			&payment.Description,
			&payment.Status,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		); err != nil {
			return nil, err
		}

		payments = append(payments, &payment)
	}

	return payments, nil
}

// Update implements repository.PaymentRepository.
func (r *PaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	updQuery := sq.Update(paymentsTable).
		Set("user_id", payment.UserID).
		Set("order_id", payment.OrderID).
		Set("currency", payment.Currency).
		Set("total_price", payment.TotalPrice).
		Set("status", payment.Status).
		Set("payment_method", payment.PaymentMethod).
		Set("description", payment.Description).
		Set("status", payment.Status).
		Set("created_at", payment.CreatedAt).
		Set("updated_at", payment.UpdatedAt).
		Where(sq.Eq{"id": payment.ID}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := updQuery.ToSql()
	if err != nil {
		return err
	}

	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

// Delete implements repository.PaymentRepository.
func (r *PaymentRepository) Delete(ctx context.Context, paymentId string) error {
	delQuery := sq.Delete(paymentsTable).
		Where(sq.Eq{"id": paymentId}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := delQuery.ToSql()
	if err != nil {
		return err
	}

	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return err
	}

	return nil
}
