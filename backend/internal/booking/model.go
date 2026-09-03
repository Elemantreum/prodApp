package booking

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type BookingStatus int

const (
	Active BookingStatus = iota
	Cancelled
	Completed
	NotCompleted
)

func (s BookingStatus) String() string {
	switch s {
	case Active:
		return "active"
	case Cancelled:
		return "cancelled"
	case Completed:
		return "completed"
	case NotCompleted:
		return "not_completed"
	default:
		return "active"
	}
}

// Value implements driver.Valuer so pgx can write a BookingStatus into the
// "status" varchar column.
func (s BookingStatus) Value() (driver.Value, error) {
	return s.String(), nil
}

// Scan implements sql.Scanner so pgx can read the "status" varchar column
// back into a BookingStatus.
func (s *BookingStatus) Scan(value any) error {
	str, err := bookingStatusStringValue(value)
	if err != nil {
		return err
	}
	return s.parse(str)
}

// MarshalJSON implements json.Marshaler so the API returns the status as a
// readable string ("active") instead of the underlying int.
func (s BookingStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements json.Unmarshaler, the mirror of MarshalJSON.
func (s *BookingStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	return s.parse(str)
}

func (s *BookingStatus) parse(str string) error {
	switch str {
	case "active":
		*s = Active
	case "cancelled":
		*s = Cancelled
	case "completed":
		*s = Completed
	case "not_completed":
		*s = NotCompleted
	default:
		return fmt.Errorf("unknown booking status: %q", str)
	}

	return nil
}

func bookingStatusStringValue(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		return "", fmt.Errorf("cannot scan %T into BookingStatus", value)
	}
}

type Booking struct {
	Id            int           `json:"id"`
	UserId        int           `json:"userId"`
	TrainingId    int           `json:"trainingId"`
	Status        BookingStatus `json:"status"`
	BookingDate   time.Time     `json:"bookingDate"`
	UnbookingDate *time.Time    `json:"unbookingDate"`
}
