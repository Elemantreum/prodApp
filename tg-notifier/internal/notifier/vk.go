package notifier

import "context"

type VkNotifier struct {
}

func (v VkNotifier) NotifyUsers(ctx context.Context, chatId []int, msg string) error {

	panic("implement me")
}

func (v VkNotifier) NotifyAll(ctx context.Context, msg string) error {

	panic("implement me")
}

func (v VkNotifier) NotifyAdmins(ctx context.Context, msg string) error {

	panic("implement me")
}

func NewVkNotifier(appSecretKey string) *VkNotifier {
	return &VkNotifier{}
}
