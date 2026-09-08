package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrRefreshTokenInvalid — токен не найден, уже отозван (использован при
// ротации/логауте) или истёк. Не различаем эти случаи в ответе клиенту —
// в обоих случаях реакция одна: "залогинься заново".
var ErrRefreshTokenInvalid = errors.New("refresh token invalid or expired")

// RefreshTokenRepository — отдельно от TokenIssuer, потому что этот токен,
// в отличие от access-JWT, opaque и должен уметь отзываться сервером
// (logout, ротация, компрометация) — а значит обязан жить в БД, а не
// проверяться одной лишь подписью.
type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

// Create генерирует новый случайный refresh token, кладёт в БД только его
// хэш (утечка БД не должна давать готовый к использованию токен) и
// возвращает клиенту сырое значение — оно нигде, кроме этого return'а, не
// сохраняется.
func (r *RefreshTokenRepository) Create(ctx context.Context, userID int, ttl time.Duration) (string, error) {
	raw, err := generateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, hashRefreshToken(raw), time.Now().Add(ttl))
	if err != nil {
		return "", fmt.Errorf("insert refresh token: %w", err)
	}

	return raw, nil
}

// FindActiveUserID проверяет, что токен существует, не отозван и не истёк,
// и отдаёт id пользователя, которому он принадлежит.
func (r *RefreshTokenRepository) FindActiveUserID(ctx context.Context, raw string) (int, error) {
	var userID int
	err := r.pool.QueryRow(ctx,
		`SELECT user_id FROM refresh_tokens
		 WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()`,
		hashRefreshToken(raw)).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrRefreshTokenInvalid
	}
	if err != nil {
		return 0, fmt.Errorf("query refresh token: %w", err)
	}
	return userID, nil
}

// Revoke гасит токен — вызывается и при ротации (старый токен гасится сразу
// после того, как по нему выдали новую пару, чтобы повторное предъявление
// украденного refresh token'а больше не срабатывало), и при явном логауте.
// Отзыв уже отозванного/несуществующего токена — не ошибка, идемпотентно.
func (r *RefreshTokenRepository) Revoke(ctx context.Context, raw string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`,
		hashRefreshToken(raw))
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAllForUser гасит все активные токены пользователя разом — на
// сейчас нигде не вызывается, но нужна для будущего "выйти со всех
// устройств"/принудительного логаута при компрометации аккаунта.
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`,
		userID)
	if err != nil {
		return fmt.Errorf("revoke refresh tokens for user %d: %w", userID, err)
	}
	return nil
}

func hashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func generateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
