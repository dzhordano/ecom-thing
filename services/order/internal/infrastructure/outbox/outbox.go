package outbox

import (
	"context"
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/kafka"
	"github.com/dzhordano/ecom-thing/services/order/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type OutboxProcessor struct {
	log      logger.BaseLogger
	db       *pgxpool.Pool
	prod     kafka.OrdersProducer
	interval time.Duration
}

type OutboxMessage struct {
	ID         string
	OrderID    string
	Topic      string
	EventType  string
	Currency   string
	TotalPrice float64
	CreatedAt  time.Time
}

func NewOutboxProcessor(log logger.BaseLogger, db *pgxpool.Pool, prod kafka.OrdersProducer, interval time.Duration) *OutboxProcessor {
	return &OutboxProcessor{
		log:      log,
		db:       db,
		prod:     prod,
		interval: interval,
	}
}

// Start запускает воркер в отдельной горутине
func (op *OutboxProcessor) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(op.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				op.processOutbox(ctx)
			case <-ctx.Done():
				op.log.Info("outbox processor shutting down")
				return
			}
		}
	}()
}

// processOutbox читает события из Outbox и публикует их в Kafka
func (op *OutboxProcessor) processOutbox(ctx context.Context) {
	tx, err := op.db.Begin(ctx)
	if err != nil {
		op.log.Error("failed to start transaction", zap.Error(err))
		return
	}

	rows, err := tx.Query(ctx,
		`SELECT id, order_id, topic, event_type, currency, total_price, created_at 
		FROM outbox 
		WHERE processed_at IS NULL 
		ORDER BY created_at ASC LIMIT 10`)
	if err != nil {
		op.log.Error("failed to query outbox", zap.Error(err))
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			op.log.Error("failed to rollback outbox transaction", zap.Error(rbErr))
		}
		return
	}

	var messages []OutboxMessage
	for rows.Next() {
		var msg OutboxMessage
		if err := rows.Scan(&msg.ID, &msg.OrderID, &msg.Topic, &msg.EventType, &msg.Currency, &msg.TotalPrice, &msg.CreatedAt); err != nil {
			op.log.Error("failed to scan outbox row", zap.Error(err))
			continue
		}
		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			op.log.Error("failed to rollback outbox transaction", zap.Error(rbErr))
		}
		return
	}

	for _, msg := range messages {

		err = op.prod.Produce(msg.Topic, -1, map[string]any{
			"order_id":    msg.OrderID,
			"currency":    msg.Currency,
			"total_price": msg.TotalPrice,
		})
		if err != nil {
			op.log.Error("failed to send Kafka message", zap.Error(err))
			continue
		}

		// Обновляем запись в Outbox, помечая как обработанную
		_, err = tx.Exec(ctx, `UPDATE outbox SET processed_at = NOW() WHERE id = $1`, msg.ID)
		if err != nil {
			op.log.Error("failed to update outbox", zap.Error(err))
			continue
		}
	}

	if err := tx.Commit(ctx); err != nil {
		op.log.Error("failed to commit outbox transaction", zap.Error(err))
	}
}
