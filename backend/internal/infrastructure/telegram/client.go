package telegram

import (
	"context"
	"prodApp/proto/gen/notifierpb"
)

type Client struct {
	client notifierpb.NotifierServiceClient
}

func NewClient(client notifierpb.NotifierServiceClient) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) Notify(
	ctx context.Context,
	recipientID int64,
	msg string,
) error {
	_, err := c.client.Notify(
		ctx,
		&notifierpb.NotifyRequest{
			RecipientId: recipientID,
			Message:     msg,
		},
	)

	return err
}

func (c *Client) NotifyMany(
	ctx context.Context,
	recipientIDs []int64,
	msg string,
) error {
	_, err := c.client.NotifyMany(
		ctx,
		&notifierpb.NotifyManyRequest{
			RecipientIds: recipientIDs,
			Message:      msg,
		},
	)

	return err
}
