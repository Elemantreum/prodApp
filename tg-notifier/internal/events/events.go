// Package events — типы событий, которые публикует бэкенд. Зеркалит
// training.OnCreateTraining/OnUpdateTraining и user.OnNewUser из монолита —
// общего Go-модуля между сервисами нет, синхронизируем руками при
// добавлении нового события с обеих сторон.
package events

import "time"

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

type OnNewUser struct {
	UserId     int
	FirstName  string
	LastName   string
	Provider   string
	ExternalId string
}
