package booking

import (
	"prodApp/internal/user"
	"time"
)

type BookingStatus string

const (
	Active       BookingStatus = "active"
	Cancelled    BookingStatus = "cancelled"
	Completed    BookingStatus = "completed"
	NotCompleted BookingStatus = "not_completed"
)

type Booking struct {
	Id            int           `json:"id"`
	UserId        int           `json:"userId"`
	TrainingId    int           `json:"trainingId"`
	Status        BookingStatus `json:"status"`
	BookingDate   time.Time     `json:"bookingDate"`
	UnbookingDate *time.Time    `json:"unbookingDate"`
}

type Participant struct {
	UserId      int       `json:"userId"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	BookingDate time.Time `json:"bookingDate"`

	Identities []user.Identity `json:"identities"`
}

type BookingDetail struct {
	BookingId   int             `json:"bookingId"`
	UserId      int             `json:"userId"`
	FirstName   string          `json:"firstName"`
	LastName    string          `json:"lastName"`
	Status      BookingStatus   `json:"status"`
	BookingDate time.Time       `json:"bookingDate"`
	Identities  []user.Identity `json:"identities"`
}

type SinnerStats struct {
	UserId            int    `json:"userId"`
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	LateCancellations int    `json:"lateCancellations"`
	NoShows           int    `json:"noShows"`
	PenaltyPoints     int    `json:"penaltyPoints"`
}

type BookingChange struct {
	PreviousUserID int    `json:"previousUserId"`
	Action         string `json:"action"`
	TrainingID     int    `json:"trainingId"`
	BookingID      int    `json:"bookingId"`
	UserID         int    `json:"userId"`
	TargetID       int    `json:"targetId"`
}

type WaitingMember struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type HistoryItem struct {
	ID          int           `json:"id"`
	UserID      int           `json:"userId"`
	Name        string        `json:"name"`
	TrainingID  int           `json:"trainingId"`
	Description string        `json:"description"`
	Date        time.Time     `json:"date"`
	Status      BookingStatus `json:"status"`
	Section     string        `json:"section"`
}
