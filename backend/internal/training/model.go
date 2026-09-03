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
}
