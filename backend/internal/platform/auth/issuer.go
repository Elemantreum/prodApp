package auth

import (
	"errors"
	"fmt"
	http2 "prodApp/internal/platform/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenIssuer struct {
	signingKey      []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewTokenIssuer(signingKey string) *TokenIssuer {
	return &TokenIssuer{
		signingKey: []byte(signingKey),
		// access живёт недолго и не отзываем — если утечёт, максимум
		// вреда ограничен этим окном (см. предупреждение в claims.go).
		accessTokenTTL: 15 * time.Minute,
		// refresh живёт долго, но хранится в БД как хэш и отзывается
		// (logout, ротация) — см. RefreshTokenRepository.
		refreshTokenTTL: 30 * 24 * time.Hour,
	}
}

// RefreshTokenTTL — на сколько выдавать refresh token. Публичный геттер,
// а не константа в auth-хендлере, чтобы политика жила в одном месте.
func (t *TokenIssuer) RefreshTokenTTL() time.Duration {
	return t.refreshTokenTTL
}

func (t *TokenIssuer) IssueAccessToken(userID int, isAdmin, isApproved bool) (string, error) {
	now := time.Now()
	claims := http2.Claims{
		UserID:     userID,
		IsAdmin:    isAdmin,
		IsApproved: isApproved,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.accessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(t.signingKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (t *TokenIssuer) Verify(rawToken string) (*http2.Claims, error) {
	claims := &http2.Claims{}

	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		// защита от подмены алгоритма — принимаем ТОЛЬКО HS256, никакой другой,
		// даже если атакующий пришлёт токен с alg=none или другим методом
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.signingKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
