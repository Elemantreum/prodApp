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

type OnUserApprovalChanged struct {
	UserId   int
	ChatId   int64
	Approved bool
}
