package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"github.com/google/uuid"
)

// Consumer represents a Sarama consumer group consumer with item service
type Consumer struct {
	cg           sarama.ConsumerGroup
	orderService interfaces.OrderService
	ready        chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
//
// FIXME Почему в примере НИКАКАЯ ошибка не возвращается?
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}
			log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s, headers[0].key = %v, headers[0].value = %v", string(message.Value), message.Timestamp, message.Topic, string(message.Headers[0].Key), string(message.Headers[0].Value))

			var orderEvent domain.OrderEvent
			err := json.Unmarshal(message.Value, &orderEvent)
			if err != nil {
				log.Printf("error parsing message.Value: %v", err)
				return nil
			}

			// FIXME Тут наверно все же в цикле перебрать. Хотя и хедер всего один всегда...
			eventType := string(message.Headers[0].Value)
			if eventType == "" {
				log.Println("eventType header not found")
				return nil
			}

			orderID, err := uuid.Parse(orderEvent.OrderID)
			if err != nil {
				log.Printf("error parsing order event's order id: %v", err)
				return nil
			}

			// FIXME Тут мб вы константы тоже
			switch eventType {
			case "order-cancelled":
				c.orderService.CancelOrder(session.Context(), orderID)
			case "order-completed":
				c.orderService.CompleteOrder(session.Context(), orderID)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

func NewConsumerGroup(brokers, topics []string, groupID string, orderService interfaces.OrderService) (*Consumer, error) {
	c := sarama.NewConfig()

	// TODO Настроить конфиг наверно.
	c.Consumer.Offsets.AutoCommit.Enable = false
	c.Consumer.MaxWaitTime = 500 * time.Millisecond

	cg, err := sarama.NewConsumerGroup(brokers, groupID, c)
	if err != nil {
		return nil, fmt.Errorf("error creating consumer group: %w", err)
	}

	return &Consumer{
		orderService: orderService,
		cg:           cg,
		ready:        make(chan bool),
	}, nil
}

func (c *Consumer) RunConsumer(ctx context.Context, topics []string) error {
	defer func() {
		cErr := c.cg.Close()
		if cErr != nil {
			log.Printf("error closing consumer group client: %v", cErr)
		}
	}()

	for {
		if err := c.cg.Consume(ctx, topics, c); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return fmt.Errorf("consumer group closed")
			}
			return fmt.Errorf("error consuming message: %w", err)
		}

		if ctx.Err() != nil {
			return fmt.Errorf("context error: %w", ctx.Err())
		}
	}
}
