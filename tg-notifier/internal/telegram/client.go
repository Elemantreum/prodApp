package telegram

import (
	"context"

	"github.com/mr-linch/go-tg"
)

type Client struct {
	client *tg.Client
}

func NewClient(token string) *Client {
	return &Client{
		client: tg.New(token),
	}
}

func (c *Client) Send(
	ctx context.Context,
	chatID int64,
	message string,
) error {
	_, err := c.client.
		SendMessage(tg.ChatID(chatID), message).
		Do(ctx)

	return err
}
