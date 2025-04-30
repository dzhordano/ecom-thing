package outbox

import (
	"context"
	"encoding/json"
	"github.com/dzhordano/ecom-thing/services/payment/internal/application/interfaces"
	"time"

	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
	"github.com/dzhordano/ecom-thing/services/payment/internal/infrastructure/kafka"
	"github.com/dzhordano/ecom-thing/services/payment/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxProcessor struct {
	log      logger.Logger
	db       *pgxpool.Pool
	prod     kafka.Producer
	interval time.Duration

	biller interfaces.Billing
}

type OutboxMessage struct {
	ID        string
	Topic     string
	EventType string
	Payload   []byte
	CreatedAt time.Time
}

func NewOutboxProcessor(log logger.Logger, db *pgxpool.Pool, prod kafka.Producer, interval time.Duration, biller interfaces.Billing) *OutboxProcessor {
	return &OutboxProcessor{
		log:      log,
		db:       db,
		prod:     prod,
		interval: interval,

		biller: biller,
	}
}

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
		`SELECT id, topic, event_type, payload, created_at
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
		if err := rows.Scan(&msg.ID, &msg.Topic, &msg.EventType, &msg.Payload, &msg.CreatedAt); err != nil {
			op.log.Error("failed to scan outbox row", "error", err)
			continue
		}
		messages = append(messages, msg)
	}
	defer rows.Close()

	if len(messages) == 0 {
		return
	}

	var pmtEv domain.OrderEvent
	for _, msg := range messages {
		if err = json.Unmarshal(msg.Payload, &pmtEv); err != nil {
			op.log.Error("failed to unmarshal outbox message", "error", err)
			continue
		}

		err = op.biller.NewPayment(ctx, pmtEv.Currency, pmtEv.TotalPrice, pmtEv.Description)
		if err != nil {
			op.log.Error("failed to process outbox message", "error", err)

			// FIXME Сейчас тут cancelled, а по факту, если не прошло -> failed.
			// Cancelled - обрабатывать отдельно. Оно также должно быть правильно обработано в Order сервисе.
			err = op.prod.Produce(msg.Topic, domain.PaymentCancelled.String(), msg.ID, pmtEv.OrderID)
			if err != nil {
				op.log.Error("failed to send Kafka message", "error", err)
				continue
			}
		} else {
			err = op.prod.Produce(msg.Topic, domain.PaymentCompleted.String(), msg.ID, pmtEv.OrderID)
			if err != nil {
				op.log.Error("failed to send Kafka message", "error", err)
				continue
			}
		}

		// Обновляем запись в Outbox, помечая как обработанную
		// TODO что будет если ляжет между обновлением и отправкой?
		// TODO можно ли сделать батчингом?
		_, err = op.db.Exec(ctx, `UPDATE outbox SET processed_at = NOW() WHERE id = $1`, msg.ID)
		if err != nil {
			op.log.Error("failed to update outbox", "error", err)
		}
	}
}
