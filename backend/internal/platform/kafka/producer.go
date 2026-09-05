// Package kafka — тонкая обёртка-продьюсер поверх Kafka. Публикует сырые
// доменные события как есть (см. training.OnCreateTraining, user.OnNewUser
// и т.п.) — какой текст слать и кому решает уже подписчик (notifier-service
// и любой другой будущий сервис, которому это интересно), не мы. Пакет
// намеренно leaf — ничего доменного не импортирует, работает с любым типом
// события через дженерик.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	kafkago "github.com/segmentio/kafka-go"
)

type Producer struct {
	writers map[string]*kafkago.Writer
}

// NewProducer сразу поднимает по одному Writer'у на каждый переданный топик —
// топики известны заранее (см. Topic* константы), лениво создавать нечего.
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

// Publish сериализует событие в JSON и публикует его в топик. Топик должен
// быть одним из переданных в NewProducer — иначе явная ошибка, а не молчаливый
// no-op.
func Publish[T any](ctx context.Context, p *Producer, topic string, event T) error {
	w, ok := p.writers[topic]
	if !ok {
		return fmt.Errorf("kafka: no writer configured for topic %q", topic)
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event for topic %q: %w", topic, err)
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
