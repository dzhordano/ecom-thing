package kafka

import (
	"context"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	"github.com/google/uuid"
	"github.com/sethvargo/go-retry"
)

// Consumer represents a Sarama consumer group consumer with item service
type Consumer struct {
	cg           sarama.ConsumerGroup
	orderService interfaces.OrderService
	ready        chan bool
	retryBackoff time.Duration
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	c.ready = make(chan bool)
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

			var eventType string
			for _, h := range message.Headers {
				if string(h.Key) == "event_type" {
					eventType = string(h.Value)
					break
				}
			}
			if len(eventType) == 0 {
				log.Printf("event_type header not found")
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

// Creates new consumer group.
//
// WARNING:
// Infinite retry loop when connecting to Kafka so it's BLOCKING.
func NewConsumerGroup(ctx context.Context, brokers []string, groupID string, orderService interfaces.OrderService) (*Consumer, error) {
	c := sarama.NewConfig()

	// TODO Настроить конфиг наверно.
	c.Version = sarama.MaxVersion
	c.Consumer.Return.Errors = true
	c.Consumer.Offsets.Initial = sarama.OffsetOldest

	// TODO Поменять на не бесконечный retry.
	var cg sarama.ConsumerGroup
	if err := retry.Do(
		ctx,
		retry.NewFibonacci(c.Metadata.Retry.Backoff),
		retry.RetryFunc(func(ctx context.Context) error {
			var err error
			cg, err = sarama.NewConsumerGroup(brokers, groupID, c)
			if err != nil {
				log.Printf("failed to create consumer group: %v, retrying...", err)
				return retry.RetryableError(err)
			}
			return nil
		}),
	); err != nil {
		return nil, err
	}

	return &Consumer{
		orderService: orderService,
		cg:           cg,
		ready:        make(chan bool),
		retryBackoff: 1 * time.Second,
	}, nil
}

func (c *Consumer) Start(ctx context.Context, topics []string) error {
	go func() {
		defer close(c.ready)
		defer c.cg.Close()

		// TODO аналогично, бесконечный ретрай..
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := c.cg.Consume(ctx, topics, c)
				switch err {
				case nil:
					c.retryBackoff = 1 * time.Second
				case sarama.ErrClosedConsumerGroup:
					return
				default:
					log.Printf("error reading from kafka: %v, retrying...", err)
					time.Sleep(c.retryBackoff)
					c.retryBackoff = min((c.retryBackoff*150)/100, 30*time.Second)
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
