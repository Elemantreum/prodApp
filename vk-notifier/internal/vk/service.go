package vk

import (
	"context"

	vk "github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/api/params"
)

type Notifier struct {
	client *vk.VK
}

func New(token string) *Notifier {
	return &Notifier{
		client: vk.NewVK(token),
	}
}

func (n *Notifier) Notify(
	ctx context.Context,
	chatID int64,
	msg string,
) error {
	p := params.NewMessagesSendBuilder()

	p.PeerID(int(chatID))
	p.Message(msg)

	_, err := n.client.MessagesSend(p.Params)

	return err
}
