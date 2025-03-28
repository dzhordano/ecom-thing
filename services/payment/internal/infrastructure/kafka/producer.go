package kafka

import (
	"log"
	"sync"

	"github.com/IBM/sarama"
)

type Producer interface {
	Produce(topic, eventType, key, orderId string) error
}

const (
	TopicPaymentCompleted = "payment-completed"
	TopicPaymentFailed    = "payment-failed"
	TopicPaymentCancelled = "payment-cancelled"
)

var (
	EventTypeHeaderKey = []byte("event_type")
)

type PaymentsSyncProducer struct {
	producerLock sync.Mutex
	producer     sarama.SyncProducer
}

func NewPaymentsSyncProducer(brokers []string, producerConfigurationProvider func() *sarama.Config) *PaymentsSyncProducer {
	producer, err := sarama.NewSyncProducer(brokers, producerConfigurationProvider())
	if err != nil {
		log.Printf("failed to start Sarama producer: %s\n", err)
		return nil
	}

	return &PaymentsSyncProducer{producer: producer}
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
