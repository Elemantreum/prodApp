package kafka

// Имена топиков = имена событий, публикуются 1 в 1 (см. main.go). Значения
// зеркалятся в notifier-service/internal/kafka/topics.go — синхронизируем
// руками при добавлении нового события, общего Go-модуля между сервисами
// нет.
const (
	TopicTrainingCreated = "OnCreateTraining"
	TopicTrainingUpdated = "OnUpdateTraining"
	TopicUserCreated     = "OnNewUser"
)
