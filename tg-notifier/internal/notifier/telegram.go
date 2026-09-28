package notifier

import (
	"context"
	"log"

	"github.com/mr-linch/go-tg"
)

type TelegramNotifier struct {
	client *tg.Client
}

func NewTelegramNotifier(telegramBotToken string) *TelegramNotifier {
	return &TelegramNotifier{
		client: tg.New(telegramBotToken),
	}
}

func (t *TelegramNotifier) NotifyMany(ctx context.Context, chatId []int64, msg string) error {
	for _, id := range chatId {
		_, err := t.client.SendMessage(tg.ChatID(id), msg).Do(ctx)
		if err != nil {
			log.Printf("notify chat %d: %v", id, err)
		}
	}

	return nil
}

func (t *TelegramNotifier) Notify(ctx context.Context, chatId int64, msg string) error {
	_, err := t.client.SendMessage(tg.ChatID(chatId), msg).Do(ctx)

	return err
}
