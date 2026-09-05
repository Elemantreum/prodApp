package kafka

// Имена топиков = имена событий, зеркалит
// backend/internal/platform/kafka/topics.go — значения должны совпадать
// 1 в 1 (это и есть ключ маршрутизации между сервисами), общего Go-модуля
// между сервисами нет, синхронизируем руками.
const (
	TopicTrainingCreated = "OnCreateTraining"
	TopicTrainingUpdated = "OnUpdateTraining"
	TopicUserCreated     = "OnNewUser"
)
