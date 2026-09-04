package notifier

import "context"

type Notifier interface {
	Notify(ctx context.Context, chatId int, msg string) error
	NotifyMany(ctx context.Context, chatId []int, msg string) error
	//NotifyAll(ctx context.Context, msg string) error
	//NotifyAdmins(ctx context.Context, msg string) error
}
