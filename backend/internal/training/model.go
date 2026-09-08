package training

import "time"

type Training struct {
	Id          int           `json:"id"`
	Description string        `json:"description"`
	Date        time.Time     `json:"date" format:"date-time" example:"2002-10-02T15:00:00Z"`
	MaxQuantity int           `json:"maxQuantity"`
	Duration    time.Duration `json:"duration" swaggertype:"primitive,integer"`
	IsActive    bool          `json:"isActive"`
	CreatedAt   time.Time     `json:"createdAt" format:"date-time" example:"2002-10-02T15:00:00Z"`
	// BookedCount — сколько активных броней сейчас на этой тренировке.
	// Считается прямым SQL-джойном на таблицу bookings прямо в репозитории
	// (см. PostgresTrainingRepository) — так дешевле, чем гонять по одному
	// GetParticipants на каждую карточку с фронта. Пакет booking сюда
	// намеренно не импортируется: booking уже импортирует training, так что
	// обратная зависимость была бы циклом.
	BookedCount int `json:"bookedCount"`
	// RecurringTrainingId — не nil, если тренировка сгенерирована правилом
	// регулярности (см. RecurringTraining). Чисто информационная метка для
	// фронта ("повторяется"), на бронирование и редактирование не влияет —
	// сгенерированную тренировку можно менять/деактивировать как обычную.
	RecurringTrainingId *int `json:"recurringTrainingId,omitempty"`
}

// RecurringTraining — правило "эта тренировка повторяется каждую неделю".
// AnchorDate — дата и время первого занятия; каждое следующее — ровно через
// 7 дней от предыдущего, без учёта дня недели/таймзоны на сервере (админ
// уже ввёл нужное локальное время при создании, дальше просто фиксированный
// недельный шаг). Пока IsActive — GenerateOccurrences материализует в
// trainings те occurrences, чьё окно создания уже наступило (см. LeadTime).
type RecurringTraining struct {
	Id          int           `json:"id"`
	Description string        `json:"description"`
	AnchorDate  time.Time     `json:"anchorDate" format:"date-time" example:"2002-10-02T15:00:00Z"`
	MaxQuantity int           `json:"maxQuantity"`
	Duration    time.Duration `json:"duration" swaggertype:"primitive,integer"`
	// LeadTime — за сколько до даты занятия его нужно создать (материализовать
	// в trainings). Например 14 дней — занятие появится в общем списке ровно
	// за две недели до начала, а не сразу на весь горизонт вперёд.
	LeadTime  time.Duration `json:"leadTime" swaggertype:"primitive,integer"`
	IsActive  bool          `json:"isActive"`
	CreatedAt time.Time     `json:"createdAt" format:"date-time" example:"2002-10-02T15:00:00Z"`
}

// OnCreateTraining/OnUpdateTraining — доменные события. Публикуются через
// outbox (см. platform/outbox — топик = имя типа события, "OnCreateTraining"/
// "OnUpdateTraining") вместо прямого похода в Kafka, так запись переживёт
// краткий даунтайм Kafka и не потеряется. Отдельный релей (main.go)
// вычитывает outbox и публикует. Внешние сервисы (tg-notifier и
// потенциально другие — например будущий сервис логирования) сами решают,
// что делать с фактом "что произошло" — здесь нет решения "кого уведомлять
// и как", это уже дело подписчика.
type OnCreateTraining struct {
	TrainingId  int
	Description string
	Date        time.Time
	MaxQuantity int
	Duration    time.Duration
}

type OnUpdateTraining struct {
	TrainingId  int
	Description string
	Date        time.Time
	MaxQuantity int
	Duration    time.Duration
	IsActive    bool
}
