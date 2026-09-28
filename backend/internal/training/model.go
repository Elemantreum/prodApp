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

	BookedCount int `json:"bookedCount"`

	RecurringTrainingId *int `json:"recurringTrainingId,omitempty"`
}

type RecurringTraining struct {
	Id          int           `json:"id"`
	Description string        `json:"description"`
	AnchorDate  time.Time     `json:"anchorDate" format:"date-time" example:"2002-10-02T15:00:00Z"`
	MaxQuantity int           `json:"maxQuantity"`
	Duration    time.Duration `json:"duration" swaggertype:"primitive,integer"`

	LeadTime  time.Duration `json:"leadTime" swaggertype:"primitive,integer"`
	IsActive  bool          `json:"isActive"`
	CreatedAt time.Time     `json:"createdAt" format:"date-time" example:"2002-10-02T15:00:00Z"`
}

type OnCreateTraining struct {
	TrainingId  int
	Description string
	Date        time.Time
	MaxQuantity int
	Duration    time.Duration
	ChatIds     []int64
}

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
