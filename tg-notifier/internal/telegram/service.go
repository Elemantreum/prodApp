package telegram

import (
	"context"

	"github.com/mr-linch/go-tg"
)

type Notifier struct {
	client *tg.Client
}

func New(token string) *Notifier {
	return &Notifier{
		client: tg.New(token),
	}
}

func (n *Notifier) Notify(
	ctx context.Context,
	chatID int64,
	msg string,
) error {
	_, err := n.client.
		SendMessage(tg.ChatID(chatID), msg).
		Do(ctx)

	return err
}
