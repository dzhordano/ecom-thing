package kafka

import (
	"context"
	"github.com/sethvargo/go-retry"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

type Producer interface {
	Produce(topic, eventType, key, orderId string) error
}

var (
	EventTypeHeaderKey = []byte("event_type")
)

type PaymentsSyncProducer struct {
	producerLock *sync.Mutex
	producer     sarama.SyncProducer
}

// NewPaymentsSyncProducer is blocking due to retries.
func NewPaymentsSyncProducer(brokers []string) (*PaymentsSyncProducer, error) {
	producerConfig := sarama.NewConfig()

	// FIXME хардкод
	producerConfig.Net.MaxOpenRequests = 1
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll
	producerConfig.Producer.Return.Successes = true

	var producer sarama.SyncProducer
	var err error
	if err = retry.Do(
		context.Background(), // Need to pass outer context here.
		retry.NewFibonacci(producerConfig.Metadata.Retry.Backoff), // TODO: fix, Using kafka default for now.
		func(ctx context.Context) error {
			producer, err = sarama.NewSyncProducer(brokers, producerConfig)
			if err != nil {
				log.Printf("failed to start Sarama producer: %s\n", err)
				return retry.RetryableError(err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &PaymentsSyncProducer{
		producerLock: &sync.Mutex{},
		producer:     producer,
	}, nil
}

func (p *PaymentsSyncProducer) Produce(topic, eventType, key, orderId string) error {
	p.producerLock.Lock()
	defer p.producerLock.Unlock()

	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Headers: []sarama.RecordHeader{
			{
				Key:   EventTypeHeaderKey,
				Value: []byte(eventType),
			},
		},
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(orderId),
	})

	return err
}

func (p *PaymentsSyncProducer) Close() error {
	return p.producer.Close()
}
