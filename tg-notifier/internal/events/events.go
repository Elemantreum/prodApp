// Package events — типы событий, которые публикует бэкенд. Зеркалит
// training.OnCreateTraining/OnUpdateTraining и user.OnNewUser/OnUserApproved
// из монолита — общего Go-модуля между сервисами нет, синхронизируем руками
// при добавлении нового события с обеих сторон.
//
// ChatIds/ChatId уже зарезолвлены бэкендом на момент публикации — этот
// сервис больше не имеет своего доступа к Postgres (раньше резолвил
// получателей сам через internal/repository.Recipients) и просто рассылает
// по готовому списку. См. training.ChatIdResolver и
// user.UserService.publishUserApproved на стороне бэкенда.
package events

import "time"

type OnCreateTraining struct {
	TrainingId  int
	Description string
	Date        time.Time
	MaxQuantity int
	Duration    time.Duration
	ChatIds     []int64
}

// OnUpdateTraining несёт и новое, и старое значение каждого поля — этот
// сервис сам строит diff для сообщения ("время: 18:00 → 19:00"), см.
// cmd/notifier/main.go.
type OnUpdateTraining struct {
	TrainingId int
	ChatIds    []int64

	Description    string
	OldDescription string
	Date           time.Time
	OldDate        time.Time
	MaxQuantity    int
	OldMaxQuantity int
	Duration       time.Duration
	OldDuration    time.Duration
	IsActive       bool
	OldIsActive    bool
}

type OnNewUser struct {
	UserId     int
	FirstName  string
	LastName   string
	Provider   string
	ExternalId string
	ChatIds    []int64
}

// OnUserApprovalChanged публикуется и при одобрении, и при снятии одобрения
// — адресовано САМОМУ юзеру, не админам. ChatId, а не ChatIds — получатель
// всегда один. Бэкенд не публикует это событие вовсе, если у юзера нет
// привязанного Telegram (см. UserService.publishApprovalChanged).
type OnUserApprovalChanged struct {
	UserId   int
	ChatId   int64
	Approved bool
}
