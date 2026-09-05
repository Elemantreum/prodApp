package notifier

import "context"

type Sender interface {
	Notify(ctx context.Context, recipientID int64, msg string) error
	NotifyMany(ctx context.Context, recipientIDs []int64, msg string) error
}
