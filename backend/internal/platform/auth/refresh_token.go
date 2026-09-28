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

var ErrRefreshTokenInvalid = errors.New("refresh token invalid or expired")

type RefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

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

func (r *RefreshTokenRepository) Revoke(ctx context.Context, raw string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`,
		hashRefreshToken(raw))
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

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
