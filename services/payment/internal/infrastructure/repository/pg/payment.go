package pg

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	paymentsTable = "payments"
	outboxTable   = "outbox"

	kafkaPaymentEvents = "payment-events"

	kafkaPaymentCreatedEvent = "created"
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
	const op = "repository.PaymentRepository.Save"

	return r.withTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		insQuery := sq.Insert(paymentsTable).
			Columns("id", "user_id", "order_id", "currency", "total_price", "status", "payment_method", "description", "created_at", "updated_at").
			Values(payment.ID, payment.UserID, payment.OrderID, payment.Currency, payment.TotalPrice, payment.Status, payment.PaymentMethod, payment.Description, payment.CreatedAt, payment.UpdatedAt).
			Suffix("RETURNING id").
			PlaceholderFormat(sq.Dollar)

		query, args, err := insQuery.ToSql()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if _, err := tx.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == pgerrcode.UniqueViolation {
					return fmt.Errorf("%s: %w", op, domain.ErrPaymentAlreadyExists)
				}
			}
			return fmt.Errorf("%s: %w", op, err)
		}

		insQuery = sq.Insert(outboxTable).
			Columns("topic", "event_type", "payload", "created_at").
			Values(kafkaPaymentEvents, kafkaPaymentCreatedEvent, payment.OrderEvent(), payment.CreatedAt).
			PlaceholderFormat(sq.Dollar)

		query, args, err = insQuery.ToSql()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if _, err := tx.Exec(ctx, query, args...); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
}

// GetById implements repository.PaymentRepository.
func (r *PaymentRepository) GetById(ctx context.Context, paymentId string, userId string) (*domain.Payment, error) {
	const op = "repository.PaymentRepository.GetById"

	selQuery := sq.Select("id", "user_id", "order_id", "currency", "total_price", "status", "payment_method", "description", "status", "created_at", "updated_at").
		From(paymentsTable).
		Where(sq.Eq{"id": paymentId, "user_id": userId}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := selQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, domain.ErrPaymentNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &payment, nil
}

// ListByUser implements repository.PaymentRepository.
func (r *PaymentRepository) ListByUser(ctx context.Context, userId string, limit uint64, offset uint64) ([]*domain.Payment, error) {
	const op = "repository.PaymentRepository.ListByUser"

	selQuery := sq.Select("id", "user_id", "order_id", "currency", "total_price", "status", "payment_method", "description", "status", "created_at", "updated_at").
		From(paymentsTable).
		Where(sq.Eq{"user_id": userId}).
		Limit(limit).
		Offset(offset).
		PlaceholderFormat(sq.Dollar)

	query, args, err := selQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var payments []*domain.Payment
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
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
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		payments = append(payments, &payment)
	}

	return payments, nil
}

// Update implements repository.PaymentRepository.
func (r *PaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	const op = "repository.PaymentRepository.Update"

	return r.withTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		updQuery := sq.Update(paymentsTable).
			Set("user_id", payment.UserID).
			Set("order_id", payment.OrderID).
			Set("currency", payment.Currency).
			Set("total_price", payment.TotalPrice).
			Set("status", payment.Status).
			Set("payment_method", payment.PaymentMethod).
			Set("description", payment.Description).
			Set("created_at", payment.CreatedAt).
			Set("updated_at", payment.UpdatedAt).
			Where(sq.Eq{"id": payment.ID}).
			PlaceholderFormat(sq.Dollar)

		query, args, err := updQuery.ToSql()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		insQuery := sq.Insert(outboxTable).
			Columns("topic", "event_type", "payload", "created_at").
			Values(kafkaPaymentEvents, kafkaPaymentCreatedEvent, payment.OrderEvent(), payment.UpdatedAt).
			PlaceholderFormat(sq.Dollar)

		query, args, err = insQuery.ToSql()
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if _, err := tx.Exec(ctx, query, args...); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
}

// Delete implements repository.PaymentRepository.
func (r *PaymentRepository) Delete(ctx context.Context, paymentId string) error {
	const op = "repository.PaymentRepository.Delete"

	delQuery := sq.Delete(paymentsTable).
		Where(sq.Eq{"id": paymentId}).
		PlaceholderFormat(sq.Dollar)

	query, args, err := delQuery.ToSql()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *PaymentRepository) withTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit(ctx)
}
