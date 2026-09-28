package kafka

import (
	"context"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"
)

type Producer struct {
	writers map[string]*kafkago.Writer
}

func NewProducer(brokers []string, topics ...string) *Producer {
	writers := make(map[string]*kafkago.Writer, len(topics))
	for _, topic := range topics {
		writers[topic] = &kafkago.Writer{
			Addr:  kafkago.TCP(brokers...),
			Topic: topic,
		}
	}
	return &Producer{writers: writers}
}

func (p *Producer) PublishRaw(ctx context.Context, topic string, payload []byte) error {
	w, ok := p.writers[topic]
	if !ok {
		return fmt.Errorf("kafka: no writer configured for topic %q", topic)
	}

	if err := w.WriteMessages(ctx, kafkago.Message{Value: payload}); err != nil {
		return fmt.Errorf("publish to topic %q: %w", topic, err)
	}

	return nil
}

func (p *Producer) Close() error {
	var firstErr error
	for _, w := range p.writers {
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
