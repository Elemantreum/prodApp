package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	kafkago "github.com/segmentio/kafka-go"
)

// Consume гоняет цикл чтения одного топика до отмены ctx: разбирает каждое
// сообщение в T и отдаёт в handle. Вся транспортная возня (подключение,
// ошибки чтения, JSON) — здесь; что событие ЗНАЧИТ и что с ним делать —
// решает handle, который каждый вызывающий пишет сам под конкретный топик
// (см. cmd/notifier/main.go). Блокирует — предполагается запуск в своей
// горутине на топик.
func Consume[T any](ctx context.Context, brokers []string, groupID, topic string, handle func(ctx context.Context, event T)) {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	defer reader.Close()

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
