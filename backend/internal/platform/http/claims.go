package http

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

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
