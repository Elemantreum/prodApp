package notifier

import "context"

// Notifier — единственный контракт этого сервиса: "отправь вот этим chat id
// вот такое сообщение". Никаких доменных типов (User, Booking, Training) —
// сервис намеренно ничего не знает про приложение целиком, только про
// telegram/vk chat id и текст. Кто уведомляется и почему — решает тот, кто
// публикует сообщение в Kafka (см. backend/internal/platform/kafka).
type Notifier interface {
	Notify(ctx context.Context, chatId int, msg string) error
	NotifyMany(ctx context.Context, chatId []int, msg string) error
}
