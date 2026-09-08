// Package kafka — тонкая обёртка-продьюсер поверх Kafka. С переходом на
// Transactional Outbox (см. platform/outbox) единственный, кто сюда
// пишет — релей в main.go: вычитывает неопубликованные записи outbox (уже
// сериализованные в JSON) и публикует их как есть. Пакет намеренно leaf —
// ничего доменного не импортирует, топики ему называют по имени явно.
package kafka

import (
	"context"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"
)

type Producer struct {
	writers map[string]*kafkago.Writer
}

// NewProducer сразу поднимает по одному Writer'у на каждый переданный топик —
// топики известны заранее (см. training.Topic*/user.Topic*), лениво
// создавать нечего.
func NewProducer(brokers []string, topics ...string) *Producer {
	writers := make(map[string]*kafkago.Writer, len(topics))
	for _, topic := range topics {
		writers[topic] = &kafkago.Writer{
			Addr:                   kafkago.TCP(brokers...),
			Topic:                  topic,
			AllowAutoTopicCreation: true,
		}
	}
	return &Producer{writers: writers}
}

// PublishRaw публикует уже сериализованный payload в топик. Топик должен
// быть одним из переданных в NewProducer — иначе явная ошибка, а не
// молчаливый no-op.
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
