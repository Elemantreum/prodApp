package user

import "time"

type User struct {
	Id         int    `json:"id"`
	Username   string `json:"username"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	IsApproved bool   `json:"isApproved"`
	IsAdmin    bool   `json:"isAdmin"`
	// IsRegistered — прошёл ли обязательный шаг "как вас зовут" при первом
	// входе. Новые аккаунты создаются с false (EnsureProfile) и переводятся в
	// true первым же успешным UpdateProfile — неважно, вызван он самим юзером
	// с экрана регистрации или админом. Чисто фронтовый гейт: на бронирование
	// и прочую бизнес-логику не влияет, в отличие от IsApproved.
	IsRegistered bool      `json:"isRegistered"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
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
