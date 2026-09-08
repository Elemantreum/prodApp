package notifier

import "context"

// VkNotifier — перенесённая заглушка, как была в backend/internal/notifier.
// Не реализована и нигде не подключена (см. cmd/notifier/main.go) — так же,
// как и в монолите до переноса.
type VkNotifier struct {
}

func (v VkNotifier) NotifyUsers(ctx context.Context, chatId []int, msg string) error {
	//TODO implement me
	panic("implement me")
}

func (v VkNotifier) NotifyAll(ctx context.Context, msg string) error {
	//TODO implement me
	panic("implement me")
}

func (v VkNotifier) NotifyAdmins(ctx context.Context, msg string) error {
	//TODO implement me
	panic("implement me")
}

func NewVkNotifier(appSecretKey string) *VkNotifier {
	return &VkNotifier{}
}
