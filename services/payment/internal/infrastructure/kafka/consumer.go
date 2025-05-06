package kafka

import (
	"context"
	"errors"
	"fmt"
	"github.com/dzhordano/ecom-thing/services/payment/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"io"
	"log"
	"math"
	"sync"
	"time"

	"github.com/dzhordano/ecom-thing/services/payment/internal/application/interfaces"
)

var (
	ErrInvalidEventType = errors.New("invalid event type")

	// To filter out unnecessary events
	events = map[string]bool{
		"order-created":   true,
		"order-cancelled": true,
	}
)

const (
	serviceGroupID = "payment-consumer-group"
)

// FIXME inject logger or use it here somehow

type Consumer struct {
	brokers      []string
	topics       []string
	ps           interfaces.PaymentService
	retryBackoff time.Duration
	retries      uint
}

// NewConsumerGroup returns new consumer group.
//
// If retries amount provided as 0, infinite (max uint) number of retries will be set
func NewConsumerGroup(ctx context.Context, brokers, topics []string, is interfaces.PaymentService, retryBackoff time.Duration, retries uint) (*Consumer, error) {
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
		ps:           is,
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
						log.Printf("error while commiting message with key: %v. error: %v\n", m.Key, err)
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
	var pmtEv domain.OrderEvent
	if err := pmtEv.UnmarshalJSON(m.Value); err != nil {
		return err
	}

	// FIXME Тут наверно все же в цикле перебрать. Хотя и хедер всего один всегда...

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

	orderID, err := uuid.Parse(pmtEv.OrderID)
	if err != nil {
		return err
	}

	userID, err := uuid.Parse(pmtEv.UserID)
	if err != nil {
		return err
	}

	// FIXME тут тоже в константы наверно. подумать об аггрегации таких событий??? (но куда :*)
	switch eventType {
	case "order-created":
		if _, err = c.ps.CreatePayment(ctx, dto.CreatePaymentRequest{
			OrderId:       orderID,
			UserId:        userID,
			Currency:      pmtEv.Currency,
			TotalPrice:    pmtEv.TotalPrice,
			PaymentMethod: pmtEv.PaymentMethod,
			Description:   pmtEv.Description,
			RedirectURL:   fmt.Sprintf("localhost:1337/payment/%s", orderID), // FIXME тут неправильно пока
		}); err != nil {
			return err
		}
	case "order-cancelled":
		if err := c.ps.CancelPayment(ctx, orderID, userID); err != nil {
			return err
		}
	}

	return nil
}
