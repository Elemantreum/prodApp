package http

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

// Claims — то, что реально кладём внутрь JWT. Живёт в platform/http (а не в
// platform/auth, откуда логически напрашивается) потому что и auth, и любой
// хендлер (booking, user, ...) должны иметь к ней доступ, а auth сам зависит
// от user (UserService) — если Claims жила бы в auth, а user читал бы её
// оттуда же, получился бы cycle auth -> user -> auth. platform/http ни от
// кого из них не зависит, так что это безопасное общее место.
//
// ВАЖНО: IsAdmin/IsApproved "замораживаются" на момент выпуска токена —
// если их поменяют в БД, уже выданный токен этого не узнает до истечения exp.
// Именно поэтому access token должен жить недолго (см. auth.TokenIssuer).
type Claims struct {
	UserID     int  `json:"user_id"`
	IsAdmin    bool `json:"is_admin"`
	IsApproved bool `json:"is_approved"`
	jwt.RegisteredClaims
}

type contextKey string

const claimsContextKey contextKey = "claims"

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}

func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}
