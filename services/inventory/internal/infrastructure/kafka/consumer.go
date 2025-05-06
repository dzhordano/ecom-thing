// TODO Поменять на не бесконечный retry.
package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"io"
	"log"
	"math"
	"sync"
	"time"

	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
)

var (
	ErrInvalidEventType = errors.New("invalid event type")

	// To filter out unnecessary events
	events = map[string]bool{
		"quantity-requested":  true,
		"quantity-released":   true,
		"quantity-subtracted": true,
	}
)

const (
	serviceGroupID = "inventory-consumer-group"
)

// FIXME inject logger or use it here somehow

type Consumer struct {
	brokers      []string
	topics       []string
	is           interfaces.ItemService
	retryBackoff time.Duration
	retries      uint
}

// NewConsumerGroup returns new consumer group.
//
// If retries amount provided as 0, infinite (max uint) number of retries will be set
func NewConsumerGroup(ctx context.Context, brokers, topics []string, is interfaces.ItemService, retryBackoff time.Duration, retries uint) (*Consumer, error) {
	if retries == 0 {
		retries = math.MaxUint
	}

	r := retries
	rb := retryBackoff
	var err error

	for ; r > 0; r-- {
		for i := range brokers {
			if ctx.Err() != nil {
				return nil, err
			}

			_, err = kafka.Dial("tcp", brokers[i])
			if err != nil {
				log.Println("error dialing broker:", brokers[i])
				continue
			}
		}
		if err != nil {
			time.Sleep(rb)
			rb = min((rb*150)/100, 30*time.Second)
			continue
		}
	}
	if err != nil {
		log.Printf("error dialing brokers: %s. error: %v", brokers, err)
		return nil, err
	}

	return &Consumer{
		brokers:      brokers,
		topics:       topics,
		is:           is,
		retryBackoff: retryBackoff,
		retries:      retries,
	}, nil
}

// RunConsumers runs amount * topics consumers. Also accepts waitgroup for graceful shutdown.
//
// If errors from service occur or commit fails, it retries.
func (c *Consumer) RunConsumers(ctx context.Context, amount int, wg *sync.WaitGroup) {
	wg.Add(amount * len(c.topics)) // Total consumers we'll be running.

	for _, topic := range c.topics {
		log.Printf("starting %d consumers for topic: %s\n", amount, topic)

		for range amount {
			go func(topic string, retries uint) {
				defer wg.Done()
				r := kafka.NewReader(kafka.ReaderConfig{
					Brokers:  c.brokers,
					GroupID:  serviceGroupID,
					Topic:    topic,
					MinBytes: 10e2, // 1  KB
					MaxBytes: 10e6, // 10 MB
				})
				defer func() {
					if err := r.Close(); err != nil {
						log.Printf("error closing consumer: %v\n", err)
					}
				}()

				for {
					m, err := r.FetchMessage(ctx)
					if err != nil {
						log.Printf("error fetching messages: %v\n", err)
						if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
							break
						}

						if retries == 0 {
							log.Printf("unable to reconnect after %d retries.\n", c.retries)
							break
						}

						log.Printf("retrying after: %v", c.retryBackoff)
						time.Sleep(c.retryBackoff)
						c.retryBackoff = min((c.retryBackoff*150)/100, 30*time.Second)
						retries--
						continue
					}

					if err := c.executeEvent(ctx, m); err != nil {
						// Retry if an event was correct. Otherwise,
						if !errors.Is(err, ErrInvalidEventType) {
							continue
						}
					}

					// FIXME здесь нет гарантии что коммит успешен, НАДО ввести ретраи + сделать записи идемпотентными. пока что risky...

					if err := r.CommitMessages(ctx, m); err != nil {
						log.Printf("error while commiting message with key: %v. err: %v\n", m.Key, err)
					}
				}
			}(topic, c.retries)
		}
	}
}

func (c *Consumer) CreateTopics(_ context.Context, partitions, replicationFactor int) (err error) {
	var conn *kafka.Conn
	for i := range c.brokers {
		conn, err = kafka.Dial("tcp", c.brokers[i])
		if err != nil {
			log.Printf("error dialing broker: %s. error: %v", c.brokers[i], err)
		}
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing kafka conn: %v\n", err)
		}
	}()

	ctrl, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("error getting controller for broker: %s. error: %v", c.brokers[0], err)
	}

	controllerAddr := fmt.Sprintf("%s:%d", ctrl.Host, ctrl.Port)
	ctrlConn, err := kafka.Dial("tcp", controllerAddr)
	if err != nil {
		return fmt.Errorf("error dialing controller: %v", err)
	}
	defer func() {
		if err := ctrlConn.Close(); err != nil {
			log.Printf("error closing controller connection: %v\n", err)
		}
	}()

	var tcfgs []kafka.TopicConfig

	for _, t := range c.topics {
		tcfgs = append(tcfgs, kafka.TopicConfig{
			Topic:             t,
			NumPartitions:     partitions,
			ReplicationFactor: replicationFactor,
		})
	}

	err = ctrlConn.CreateTopics(tcfgs...)
	if err != nil {
		return fmt.Errorf("error creating topics on controller %s. error: %v", controllerAddr, err)
	}

	return nil
}

func (c *Consumer) executeEvent(ctx context.Context, m kafka.Message) error {
	var invEvent struct {
		OrderID string `json:"order_id"`
		Items   []struct {
			ProductID string
			Quantity  uint64
		} `json:"items"`
	}

	if err := json.Unmarshal(m.Value, &invEvent); err != nil {
		return err
	}

	items := map[string]uint64{}
	for i := range invEvent.Items {
		items[invEvent.Items[i].ProductID] = invEvent.Items[i].Quantity
	}

	var eventType string
	for _, h := range m.Headers {
		if string(h.Key) == "event_type" {
			eventType = string(h.Value)
			break
		}
	}
	if !events[eventType] {
		return ErrInvalidEventType
	}

	// FIXME тут тоже в константы наверно. подумать об аггрегации таких событий??? (но куда :*)
	switch eventType {
	case "quantity-requested":
		if err := c.is.SetItemsWithOp(ctx, items, domain.OperationLock); err != nil {
			return err
		}
	case "quantity-released":
		if err := c.is.SetItemsWithOp(ctx, items, domain.OperationUnlock); err != nil {
			return err
		}
	case "quantity-subtracted":
		if err := c.is.SetItemsWithOp(ctx, items, domain.OperationSubLocked); err != nil {
			return err
		}
	}

	return nil
}
