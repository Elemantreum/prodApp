package user

import "time"

type User struct {
	Id         int       `json:"id"`
	Username   string    `json:"username"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	IsApproved bool      `json:"isApproved"`
	IsAdmin    bool      `json:"isAdmin"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Identity struct {
	Id         int
	UserId     int
	Provider   string
	ExternalId string
	CreatedAt  time.Time
}

// LinkCode — одноразовый короткоживущий код: генерируется в уже открытой
// сессии ("привяжи мне ещё и VK"), вводится при входе с другой платформы,
// чтобы обе учётки схлопнулись в одного User вместо двух разных.
type LinkCode struct {
	Code      string
	UserId    int
	ExpiresAt time.Time
}
