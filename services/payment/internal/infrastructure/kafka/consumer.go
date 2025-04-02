package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/dzhordano/ecom-thing/services/payment/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/payment/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
	"github.com/google/uuid"
	"github.com/sethvargo/go-retry"
)

// Consumer represents a Sarama consumer group consumer with item service
type Consumer struct {
	cg             sarama.ConsumerGroup
	paymentService interfaces.PaymentService
	ready          chan bool
	retryBackoff   time.Duration
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

			var pmtEv domain.OrderEvent
			err := pmtEv.UnmarshalJSON(message.Value)
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

			orderID, err := uuid.Parse(pmtEv.OrderID)
			if err != nil {
				log.Printf("error parsing order event's order id: %v", err)
				return nil
			}

			userID, err := uuid.Parse(pmtEv.UserID)
			if err != nil {
				log.Printf("error parsing order event's user id: %v", err)
				return nil
			}

			// FIXME тут тоже в константы наверно. подумать об аггрегации таких событий??? (но куда :*)
			switch eventType {
			case "order-created":
				c.paymentService.CreatePayment(session.Context(), dto.CreatePaymentRequest{
					OrderId:       orderID,
					UserId:        userID,
					Currency:      pmtEv.Currency,
					TotalPrice:    pmtEv.TotalPrice,
					PaymentMethod: pmtEv.PaymentMethod,
					Description:   pmtEv.Description,
					RedirectURL:   fmt.Sprintf("localhost:3000/payment/%s", orderID), // FIXME тут неправильно пока
				})
			case "order-cancelled":
				c.paymentService.CancelPayment(session.Context(), orderID, userID)
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
func NewConsumerGroup(ctx context.Context, brokers []string, groupID string, paymentService interfaces.PaymentService) (*Consumer, error) {
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
		paymentService: paymentService,
		cg:             cg,
		ready:          make(chan bool),
		retryBackoff:   1 * time.Second,
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
					log.Println("consumer group is closed, exiting")
					return
				default:
					log.Printf("error from consumer group: %v, retrying...", err)
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
