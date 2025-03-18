package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/IBM/sarama"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
)

// Consumer represents a Sarama consumer group consumer with item service
type Consumer struct {
	cg               sarama.ConsumerGroup
	inventoryService interfaces.ItemService
	ready            chan bool
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
			// FIXME Тут наверно все же в цикле перебрать. Хотя и хедер всего один всегда...
			eventType := string(message.Headers[0].Value)
			if eventType == "" {
				log.Printf("eventType header not found")
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

func NewConsumerGroup(brokers []string, groupID string, inventoryService interfaces.ItemService) (*Consumer, error) {
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
		inventoryService: inventoryService,
		cg:               cg,
		ready:            make(chan bool),
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
