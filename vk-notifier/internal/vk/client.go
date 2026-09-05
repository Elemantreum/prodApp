package vk

import "C"
import (
	"context"

	vk "github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/api/params"
)

type Client struct {
	client *vk.VK
}

func NewClient(token string) *Client {
	return &Client{
		client: vk.NewVK(token),
	}
}

func (c *Client) Send(
	ctx context.Context,
	chatID int64,
	message string,
) error {
	p := params.NewMessagesSendBuilder()

	p.PeerID(int(chatID))
	p.Message(message)

	_, err := c.client.MessagesSend(p.Params)

	return err
}
