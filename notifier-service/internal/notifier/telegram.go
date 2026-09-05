package notifier

import (
	"context"
	"log"

	"github.com/mr-linch/go-tg"
)

// TelegramNotifier — перенесён как есть из backend/internal/notifier
// (см. историю бэкенда): раньше жил в монолите и резолвил получателей сам
// через *user.UserService (NotifyAll/NotifyAdmins). Теперь резолв "кому"
// целиком на стороне бэкенда (он публикует уже готовый список chat id в
// Kafka) — этот сервис только шлёт, никакой зависимости от user/booking
// пакетов и, соответственно, от Postgres здесь больше нет.
type TelegramNotifier struct {
	client *tg.Client
}

func NewTelegramNotifier(telegramBotToken string) *TelegramNotifier {
	return &TelegramNotifier{
		client: tg.New(telegramBotToken),
	}
}

func (t *TelegramNotifier) NotifyMany(ctx context.Context, chatId []int, msg string) error {
	for _, id := range chatId {
		_, err := t.client.SendMessage(tg.ChatID(id), msg).Do(ctx)
		if err != nil {
			log.Printf("notify chat %d: %v", id, err)
		}
	}

	return nil
}

func (t *TelegramNotifier) Notify(ctx context.Context, chatId int, msg string) error {
	_, err := t.client.SendMessage(tg.ChatID(chatId), msg).Do(ctx)

	return err
}
