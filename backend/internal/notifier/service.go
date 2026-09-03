package notifier

import "context"

type Notifier interface {
	NotifyUser(ctx context.Context, chatID int, msg string) error
	NotifyAll(ctx context.Context, msg string) error
	NotifyAdmins(ctx context.Context, msg string) error
	NotifyTraining(ctx context.Context, training int, msg string) error
}
