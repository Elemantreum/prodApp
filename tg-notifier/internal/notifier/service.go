package notifier

import "context"

type Notifier interface {
	Notify(ctx context.Context, chatId int64, msg string) error
	NotifyMany(ctx context.Context, chatId []int64, msg string) error
}
