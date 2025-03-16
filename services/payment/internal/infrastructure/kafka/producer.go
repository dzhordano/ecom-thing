package kafka

import (
	"log"
	"sync"

	"github.com/IBM/sarama"
)

type Producer interface {
	Produce(topic string, orderId string) error
}

const (
	TopicPaymentCompleted = "payment-completed"
	TopicPaymentFailed    = "payment-failed"
	TopicPaymentCancelled = "payment-cancelled"
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

func (p *PaymentsSyncProducer) Produce(topic string, orderId string) error {
	p.producerLock.Lock()
	defer p.producerLock.Unlock()

	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(orderId),
	})

	return err
}

func (p *PaymentsSyncProducer) Close() error {
	return p.producer.Close()
}
