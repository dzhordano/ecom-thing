package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"github.com/sethvargo/go-retry"
)

// Consumer represents a Sarama consumer group consumer with item service
type Consumer struct {
	cg               sarama.ConsumerGroup
	inventoryService interfaces.ItemService
	ready            chan bool
	retryBackoff     time.Duration
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
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}

			var invEvent struct {
				OrderID string `json:"order_id"`
				Items   []struct {
					ProductID string
					Quantity  uint64
				} `json:"items"`
			}
			err := json.Unmarshal(message.Value, &invEvent)
			if err != nil {
				log.Printf("error parsing message.Value: %v", err)
				return nil
			}

			items := map[string]uint64{}
			for i := range invEvent.Items {
				items[invEvent.Items[i].ProductID] = invEvent.Items[i].Quantity
			}

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

			// FIXME тут тоже в константы наверно. подумать об аггрегации таких событий??? (но куда :*)
			switch eventType {
			case "quantity-requested":
				c.inventoryService.SetItemsWithOp(session.Context(), items, domain.OperationLock)
			case "quantity-released":
				c.inventoryService.SetItemsWithOp(session.Context(), items, domain.OperationUnlock)
			case "quantity-subtracted":
				c.inventoryService.SetItemsWithOp(session.Context(), items, domain.OperationSubLocked)
			}

			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func NewConsumerGroup(ctx context.Context, brokers []string, groupID string, inventoryService interfaces.ItemService) (*Consumer, error) {
	c := sarama.NewConfig()

	c.Version = sarama.MaxVersion
	c.Consumer.Return.Errors = true
	c.Consumer.Offsets.Initial = sarama.OffsetOldest

	// WARNING:
	// Infinite retry loop so is BLOCKING.
	var cg sarama.ConsumerGroup
	if err := retry.Do(
		ctx,
		retry.NewFibonacci(c.Metadata.Retry.Backoff),
		retry.RetryFunc(func(ctx context.Context) error {
			log.Println("attempting to create consumer group...")
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
		inventoryService: inventoryService,
		cg:               cg,
		ready:            make(chan bool),
		retryBackoff:     1 * time.Second,
	}, nil
}

func (c *Consumer) Start(ctx context.Context, topics []string) error {
	go func() {
		defer close(c.ready)
		defer c.cg.Close()

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
					log.Println("consumer group is closed, exiting")
					return
				default:
					log.Printf("error from consumer group: %v, retrying...", err)
					time.Sleep(c.retryBackoff)
					// FIXME хардкод ретрая макс времени на ожидание. Вроде норм...
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
