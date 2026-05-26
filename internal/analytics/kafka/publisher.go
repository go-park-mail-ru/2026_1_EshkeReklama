package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"eshkere/internal/analytics"

	kafkago "github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafkago.Writer
}

func NewPublisher(brokers []string, topic string) (*Publisher, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers cannot be empty")
	}
	if topic == "" {
		return nil, fmt.Errorf("kafka topic cannot be empty")
	}

	return &Publisher{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafkago.Hash{},
			RequiredAcks: kafkago.RequireOne,
			BatchTimeout: 10 * time.Millisecond,
		},
	}, nil
}

func (p *Publisher) PublishAdEvent(ctx context.Context, event analytics.AdEvent) error {
	if p == nil || p.writer == nil {
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal ad event: %w", err)
	}

	return p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(event.RequestID),
		Value: payload,
		Time:  event.OccurredAt,
	})
}

func (p *Publisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}
