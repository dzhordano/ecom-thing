package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
	"math"
	"time"
)

// TODO сделать async
// как использовать partitions?

type Producer interface {
	Produce(ctx context.Context, eventType, key string, payload []byte) error
}

var (
	EventTypeHeaderKey = "event_type"
)

type KafkaProducer struct {
	ws []*kafka.Writer

	brokers      []string
	topics       []string
	retryBackoff time.Duration
	retries      uint
}

// NewProducer creates KafkaProducer.
//
// If retries set 0, infinite (max uint) number of retries will be set.
func NewProducer(brokers []string, topics []string, retryBackoff time.Duration, retries uint) *KafkaProducer {
	if retries == 0 {
		retries = math.MaxUint
	}

	var ws []*kafka.Writer

	for _, topic := range topics {
		ws = append(ws, &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.RoundRobin{},
		})
	}

	return &KafkaProducer{ws: ws, brokers: brokers, topics: topics, retryBackoff: retryBackoff, retries: retries}
}

func (p *KafkaProducer) Produce(ctx context.Context, eventType, key string, payload []byte) error {
	m := kafka.Message{
		Key:   []byte(key),
		Value: payload,
		Headers: []kafka.Header{
			{
				Key:   EventTypeHeaderKey,
				Value: []byte(eventType),
			},
		},
	}

	for i := range p.ws {
		err := p.ws[i].WriteMessages(ctx, m)
		if err != nil {
			log.Printf("error writing message to kafka: %v\n", err)
			return err
		}
	}

	return nil
}

func (p *KafkaProducer) Close() {
	for i := range p.ws {
		if err := p.ws[i].Close(); err != nil {
			log.Printf("error closing KafkaProducer: %v\n", err)
		}
	}
}
