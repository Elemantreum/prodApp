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
