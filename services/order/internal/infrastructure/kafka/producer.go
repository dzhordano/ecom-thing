package kafka

import (
	"context"
	"github.com/sethvargo/go-retry"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

// TODO сделать async
// как использовать partitions?

type Producer interface {
	Produce(topic, eventType, key string, payload []byte) error
}

const (
	TopicOrderCreated   = "order-created"
	TopicOrderCancelled = "order-cancelled"
)

var (
	EventTypeHeaderKey = []byte("event_type")
)

type OrdersSyncProducer struct {
	producerLock sync.Mutex
	producer     sarama.SyncProducer
}

func NewOrdersSyncProducer(brokers []string) (*OrdersSyncProducer, error) {
	producerConfig := sarama.NewConfig()

	// FIXME Хардкод...
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

	return &OrdersSyncProducer{
		producerLock: sync.Mutex{},
		producer:     producer,
	}, nil
}

func (p *OrdersSyncProducer) Produce(topic, eventType, key string, payload []byte) error {
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
		Value: sarama.ByteEncoder(payload),
	})

	return err
}

func (p *OrdersSyncProducer) Close() error {
	return p.producer.Close()
}

// type topicPartition struct {
// 	topic     string
// 	partition int32
// }

// type AsyncProducerProvider struct {
// 	producersLock sync.Mutex
// 	producers     map[topicPartition][]sarama.AsyncProducer

// 	producerProvider func(topic string, partition int32) sarama.AsyncProducer
// }

// func NewProducerProvider(brokers []string, producerConfigurationProvider func() *sarama.Config) *AsyncProducerProvider {
// 	provider := &AsyncProducerProvider{
// 		producers: make(map[topicPartition][]sarama.AsyncProducer),
// 	}
// 	provider.producerProvider = func(topic string, partition int32) sarama.AsyncProducer {
// 		config := producerConfigurationProvider()
// 		if config.Producer.Transaction.ID != "" {
// 			config.Producer.Transaction.ID = config.Producer.Transaction.ID + "-" + topic + "-" + fmt.Sprint(partition)
// 		}
// 		producer, err := sarama.NewAsyncProducer(brokers, config)
// 		if err != nil {
// 			return nil
// 		}
// 		return producer
// 	}
// 	return provider
// }

// func (p *AsyncProducerProvider) borrow(topic string, partition int32) (producer sarama.AsyncProducer) {
// 	p.producersLock.Lock()
// 	defer p.producersLock.Unlock()

// 	tp := topicPartition{topic: topic, partition: partition}

// 	if producers, ok := p.producers[tp]; !ok || len(producers) == 0 {
// 		for {
// 			producer = p.producerProvider(topic, partition)
// 			if producer != nil {
// 				return
// 			}
// 		}
// 	}

// 	index := len(p.producers[tp]) - 1
// 	producer = p.producers[tp][index]
// 	p.producers[tp] = p.producers[tp][:index]
// 	return
// }

// func (p *AsyncProducerProvider) release(topic string, partition int32, producer sarama.AsyncProducer) {
// 	p.producersLock.Lock()
// 	defer p.producersLock.Unlock()

// 	if producer.TxnStatus()&sarama.ProducerTxnFlagInError != 0 {
// 		// Try to close it
// 		_ = producer.Close()
// 		return
// 	}
// 	tp := topicPartition{topic: topic, partition: partition}
// 	p.producers[tp] = append(p.producers[tp], producer)
// }

// func (p *AsyncProducerProvider) Clear() {
// 	p.producersLock.Lock()
// 	defer p.producersLock.Unlock()

// 	for _, producers := range p.producers {
// 		for _, producer := range producers {
// 			producer.Close()
// 		}
// 	}
// 	for _, producers := range p.producers {
// 		producers = producers[:0] // TODO huh?
// 	}
// }

// func (p *AsyncProducerProvider) SendMessage(topic string, partition int32, message *sarama.ProducerMessage) error {
// 	producer := p.borrow(topic, partition)

// 	err := producer.BeginTxn()
// 	if err != nil {
// 		return err
// 	}

// 	producer.Input() <- message

// 	err = producer.CommitTxn()
// 	if err != nil {
// 		return err
// 	}

// 	p.release(topic, partition, producer)

// 	return nil
// }
