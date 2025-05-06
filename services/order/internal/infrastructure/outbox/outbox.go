package outbox

import (
	"context"
	"time"

	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/kafka"
	"github.com/dzhordano/ecom-thing/services/order/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxProcessor struct {
	log      logger.Logger
	db       *pgxpool.Pool
	prod     kafka.Producer
	interval time.Duration
}

type OutboxMessage struct {
	ID        string
	Topic     string
	EventType string
	Payload   []byte
	CreatedAt time.Time
}

func NewOutboxProcessor(log logger.Logger, db *pgxpool.Pool, prod kafka.Producer, interval time.Duration) *OutboxProcessor {
	return &OutboxProcessor{
		log:      log,
		db:       db,
		prod:     prod,
		interval: interval,
	}
}

// Start запускает воркер в отдельной горутине.
//
// Воркер сам завершит работу при отмене контекста.
func (op *OutboxProcessor) Start(ctx context.Context) {
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
}

// processOutbox читает события из Outbox и публикует их в Kafka
func (op *OutboxProcessor) processOutbox(ctx context.Context) {

	rows, err := op.db.Query(ctx,
		`SELECT id, event_type, payload, created_at 
		FROM outbox 
		WHERE processed_at IS NULL 
		ORDER BY created_at ASC LIMIT 10`) // TODO магич число 10
	if err != nil {
		op.log.Error("failed to query outbox", "error", err)
		return
	}

	var messages []OutboxMessage
	for rows.Next() {
		var msg OutboxMessage
		if err := rows.Scan(&msg.ID, &msg.EventType, &msg.Payload, &msg.CreatedAt); err != nil {
			op.log.Error("failed to scan outbox row", "error", err)
			continue
		}
		messages = append(messages, msg)
	}
	defer rows.Close()

	if len(messages) == 0 {
		return
	}

	for _, msg := range messages {
		err = op.prod.Produce(ctx, msg.EventType, msg.ID, msg.Payload)
		if err != nil {
			op.log.Error("failed to send Kafka message", "error", err)
			continue
		}

		// Обновляем запись в Outbox, помечая как обработанную
		// TODO можно ли сделать батчингом?
		_, err = op.db.Exec(ctx, `UPDATE outbox SET processed_at = NOW() WHERE id = $1`, msg.ID)
		if err != nil {
			op.log.Error("failed to update outbox", "error", err)
		}
	}
}
