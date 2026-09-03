package kafkaadapter

import (
	"context"
	"errors"

	segmentio "github.com/segmentio/kafka-go"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
)

type Publisher struct {
	writer *segmentio.Writer
}

func NewPublisher(brokers []string) *Publisher {
	if len(brokers) == 0 {
		panic("kafka adapter: no brokers")
	}
	return &Publisher{writer: &segmentio.Writer{
		Addr:         segmentio.TCP(brokers...),
		Balancer:     &segmentio.Hash{},
		RequiredAcks: segmentio.RequireAll,
		Async:        false,
	}}
}

func (p *Publisher) Publish(ctx context.Context, message application.OutboxMessage) error {
	return p.writer.WriteMessages(ctx, segmentio.Message{
		Topic: message.Topic,
		Key:   []byte(message.Key),
		Value: append([]byte(nil), message.Payload...),
		Headers: []segmentio.Header{
			{Key: "event_id", Value: []byte(message.ID)},
			{Key: "content-type", Value: []byte("application/json")},
		},
		Time: message.CreatedAt,
	})
}

func (p *Publisher) Close() error {
	if p == nil || p.writer == nil {
		return errors.New("kafka adapter: publisher is not initialized")
	}
	return p.writer.Close()
}
