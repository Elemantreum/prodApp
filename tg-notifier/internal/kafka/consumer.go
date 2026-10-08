package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"reflect"

	kafkago "github.com/segmentio/kafka-go"
)

func Consume[T any](ctx context.Context, brokers []string, groupID string, handle func(ctx context.Context, event T)) {
	topic := reflect.TypeFor[T]().Name()

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("%s: close reader: %v", topic, err)
		}
	}()

	log.Printf("tg-notifier: listening on topic %q", topic)

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Printf("%s: read message: %v", topic, err)
			continue
		}

		var event T
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("%s: unmarshal event: %v", topic, err)
			continue
		}

		handle(ctx, event)
	}
}
