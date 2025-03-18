package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
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

			orderIdValue := string(message.Value)

			// FIXME Тут наверно все же в цикле перебрать. Хотя и хедер всего один всегда...
			eventType := string(message.Headers[0].Value)
			if eventType == "" {
				log.Println("eventType header not found")
				return nil
			}

			orderID, err := uuid.Parse(orderIdValue)
			if err != nil {
				log.Printf("error parsing order event's order id: %v", err)
				return nil
			}

			// FIXME Тут мб константы тоже
			switch eventType {
			case "cancelled":
				c.orderService.CancelOrder(session.Context(), orderID)
			case "completed":
				c.orderService.CompleteOrder(session.Context(), orderID)
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil // TODO тут правильно?
		}
	}
}

func NewConsumerGroup(brokers []string, groupID string, orderService interfaces.OrderService) (*Consumer, error) {
	c := sarama.NewConfig()

	// TODO Настроить конфиг наверно.
	c.Version = sarama.MaxVersion
	c.Consumer.Return.Errors = true
	c.Consumer.Offsets.Initial = sarama.OffsetOldest

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

func (c *Consumer) Start(ctx context.Context, topics []string) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := c.cg.Consume(ctx, topics, c); err != nil {
					log.Printf("consuming error: %v", err)
				}
				if ctx.Err() != nil {
					return
				}
			}
		}
	}()

	<-c.ready
	return nil
}
