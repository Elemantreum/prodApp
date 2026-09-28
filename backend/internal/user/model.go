package user

import "time"

type User struct {
	Id         int    `json:"id"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	IsApproved bool   `json:"isApproved"`
	IsAdmin    bool   `json:"isAdmin"`

	IsRegistered bool      `json:"isRegistered"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type Identity struct {
	Id         int    `json:"id"`
	UserId     int    `json:"userId"`
	Provider   string `json:"provider"`
	ExternalId string `json:"externalId"`

	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
}

type LinkCode struct {
	Code      string
	UserId    int
	ExpiresAt time.Time
}
