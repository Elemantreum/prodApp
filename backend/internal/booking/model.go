package booking

import (
	"prodApp/internal/user"
	"time"
)

type BookingStatus string

const (
	Active       BookingStatus = "active"
	Cancelled                  = "cancelled"
	Completed                  = "completed"
	NotCompleted               = "not_completed"
)

type Booking struct {
	Id            int           `json:"id"`
	UserId        int           `json:"userId"`
	TrainingId    int           `json:"trainingId"`
	Status        BookingStatus `json:"status"`
	BookingDate   time.Time     `json:"bookingDate"`
	UnbookingDate *time.Time    `json:"unbookingDate"`
}

// Participant — то, что видит обычный пользователь про других записавшихся:
// имя, ник и когда записался. Никакого approve/admin статуса — это не его
// дело, в отличие от админской ручки /user/get.
type Participant struct {
	UserId      int       `json:"userId"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Username    string    `json:"username"`
	BookingDate time.Time `json:"bookingDate"`
	// Identities — привязанные платформы участника (провайдер, внешний id,
	// @-хендл если есть) — фронт строит из них ссылки на vk.com/id.../t.me/... .
	Identities []user.Identity `json:"identities"`
}

// BookingDetail — админский взгляд на одну запись на тренировку: в отличие
// от Participant включает статус и id самой брони (нужен, чтобы вызвать
// SetAttendance), и не фильтрует по статусу — сюда попадают и отменённые.
type BookingDetail struct {
	BookingId   int             `json:"bookingId"`
	UserId      int             `json:"userId"`
	FirstName   string          `json:"firstName"`
	LastName    string          `json:"lastName"`
	Username    string          `json:"username"`
	Status      BookingStatus   `json:"status"`
	BookingDate time.Time       `json:"bookingDate"`
	Identities  []user.Identity `json:"identities"`
}
