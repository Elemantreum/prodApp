package user

import (
	"context"
	"errors"
	"fmt"
	"prodApp/backend/internal/platform/apperrors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresIdentityRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresIdentityRepository(pool *pgxpool.Pool) *PostgresIdentityRepository {
	return &PostgresIdentityRepository{pool: pool}
}

func (r *PostgresIdentityRepository) GetBatchById(ctx context.Context, ids []int, provider string) ([]Identity, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, provider, external_id, username
		 FROM user_identities
		 WHERE provider = $1`, provider)
	if err != nil {
		return nil, fmt.Errorf("GetBatchById: %w", err)
	}
	defer rows.Close()

	identities := make([]Identity, 0)
	for rows.Next() {
		var identity Identity
		if err := rows.Scan(&identity.Id, &identity.UserId, &identity.Provider, &identity.ExternalId, &identity.Username); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		identities = append(identities, identity)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return identities, err
}

func (r *PostgresIdentityRepository) FindByProviderAndExternalID(ctx context.Context, provider, externalID string) (*Identity, error) {
	var identity Identity
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, provider, external_id, username
		 FROM user_identities
		 WHERE provider = $1 AND external_id = $2`,
		provider, externalID,
	).Scan(&identity.Id, &identity.UserId, &identity.Provider, &identity.ExternalId, &identity.Username)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("identity %s/%s: %w", provider, externalID, apperrors.NotFoundError{})
	}
	if err != nil {
		return nil, fmt.Errorf("find identity: %w", err)
	}
	return &identity, nil
}

func (r *PostgresIdentityRepository) Create(ctx context.Context, identity *Identity) (*Identity, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO user_identities (user_id, provider, external_id, username)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, provider, external_id, username`,
		identity.UserId, identity.Provider, identity.ExternalId, identity.Username,
	)

	var saved Identity
	err := row.Scan(&saved.Id, &saved.UserId, &saved.Provider, &saved.ExternalId, &saved.Username)
	if err != nil {
		return nil, fmt.Errorf("create identity: %w", err)
	}
	return &saved, nil
}

func (r *PostgresIdentityRepository) FindByUserID(ctx context.Context, userID int) ([]Identity, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, provider, external_id, username FROM user_identities WHERE user_id = $1 ORDER BY id`, userID)
	if err != nil {
		return nil, fmt.Errorf("find identities: %w", err)
	}
	defer rows.Close()

	identities := make([]Identity, 0)
	for rows.Next() {
		var identity Identity
		if err := rows.Scan(&identity.Id, &identity.UserId, &identity.Provider, &identity.ExternalId, &identity.Username); err != nil {
			return nil, fmt.Errorf("scan identity: %w", err)
		}
		identities = append(identities, identity)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate identities: %w", err)
	}

	return identities, nil
}

// Reassign — upsert поверх UNIQUE(provider, external_id): если такая
// идентичность уже привязана (например, юзер открыл вторую платформу без
// кода и она автосоздала себе отдельный аккаунт), просто переставляет её на
// newUserID вместо отказа. Используется ТОЛЬКО во флоу погашения кода
// привязки — там владение обеими платформенными аккаунтами уже подтверждено
// (валидная подпись initData/launch params + код, который знает только
// владелец исходного аккаунта).
func (r *PostgresIdentityRepository) Reassign(ctx context.Context, provider, externalID, username string, newUserID int) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO user_identities (user_id, provider, external_id, username)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (provider, external_id) DO UPDATE SET user_id = EXCLUDED.user_id, username = EXCLUDED.username`,
		newUserID, provider, externalID, username,
	)
	if err != nil {
		return fmt.Errorf("reassign identity: %w", err)
	}
	return nil
}

// UpdateUsername обновляет @-хендл при повторном логине — в отличие от
// имени/фамилии это не то, чем управляет юзер, а просто техническая деталь
// для ссылок, поэтому обновляем молча на каждый вход без всякого UI.
func (r *PostgresIdentityRepository) UpdateUsername(ctx context.Context, provider, externalID, username string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE user_identities SET username = $3 WHERE provider = $1 AND external_id = $2`,
		provider, externalID, username,
	)
	if err != nil {
		return fmt.Errorf("update identity username: %w", err)
	}
	return nil
}

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (p *PostgresUserRepository) Create(ctx context.Context, user *User) (*User, error) {
	row := p.pool.QueryRow(ctx,
		`INSERT INTO users (username, first_name, last_name, is_approved, is_admin, is_registered)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, username, first_name, last_name, is_approved, is_admin, is_registered, created_at, updated_at`,
		user.Username, user.FirstName, user.LastName, user.IsApproved, user.IsAdmin, user.IsRegistered,
	)

	var saved User
	err := row.Scan(&saved.Id, &saved.Username, &saved.FirstName, &saved.LastName, &saved.IsApproved, &saved.IsAdmin, &saved.IsRegistered, &saved.CreatedAt, &saved.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return &saved, nil
}

func (p *PostgresUserRepository) Update(ctx context.Context, user *User) (*User, error) {
	row := p.pool.QueryRow(ctx,
		`UPDATE users SET username = $2, first_name = $3, last_name = $4, is_approved = $5, is_admin = $6, is_registered = $7, updated_at = now()
		 WHERE id = $1
		 RETURNING id, username, first_name, last_name, is_approved, is_admin, is_registered, created_at, updated_at`,
		user.Id, user.Username, user.FirstName, user.LastName, user.IsApproved, user.IsAdmin, user.IsRegistered,
	)

	var saved User
	err := row.Scan(&saved.Id, &saved.Username, &saved.FirstName, &saved.LastName, &saved.IsApproved, &saved.IsAdmin, &saved.IsRegistered, &saved.CreatedAt, &saved.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user with id %d: %w", user.Id, apperrors.NotFoundError{})
	}
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return &saved, nil
}

func (p *PostgresUserRepository) Delete(ctx context.Context, user *User) error {
	tag, err := p.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", user.Id)
	if err != nil {
		return fmt.Errorf("delete user with id %d: %w", user.Id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user with id %d: %w", user.Id, apperrors.NotFoundError{})
	}
	return nil
}

func (p *PostgresUserRepository) FindById(ctx context.Context, id int) (*User, error) {
	var user User
	err := p.pool.QueryRow(ctx,
		`SELECT id, username, first_name, last_name, is_approved, is_admin, is_registered, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&user.Id, &user.Username, &user.FirstName, &user.LastName, &user.IsApproved, &user.IsAdmin, &user.IsRegistered, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("user with id %d: %w", id, apperrors.NotFoundError{})
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &user, nil
}

func (p *PostgresUserRepository) GetAll(ctx context.Context) ([]User, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT id, username, first_name, last_name, is_approved, is_admin, is_registered, created_at, updated_at FROM users`)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Id, &user.Username, &user.FirstName, &user.LastName, &user.IsApproved, &user.IsAdmin, &user.IsRegistered, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

func (p *PostgresUserRepository) FindAdmins(ctx context.Context) ([]User, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT id, username, first_name, last_name, is_approved, is_admin, created_at, updated_at
			FROM users
			WHERE is_admin = true
			`)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Id, &user.Username, &user.FirstName, &user.LastName, &user.IsApproved, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

type PostgresLinkCodeRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresLinkCodeRepository(pool *pgxpool.Pool) *PostgresLinkCodeRepository {
	return &PostgresLinkCodeRepository{pool: pool}
}

func (r *PostgresLinkCodeRepository) Create(ctx context.Context, lc *LinkCode) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO link_codes (code, user_id, expires_at) VALUES ($1, $2, $3)`,
		lc.Code, lc.UserId, lc.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create link code: %w", err)
	}
	return nil
}

// FindAndDelete забирает код и сразу удаляет его одним запросом — код
// одноразовый, повторное использование того же кода после успешного
// погашения (или после гонки из двух параллельных попыток) должно упасть.
func (r *PostgresLinkCodeRepository) FindAndDelete(ctx context.Context, code string) (*LinkCode, error) {
	var lc LinkCode
	err := r.pool.QueryRow(ctx,
		`DELETE FROM link_codes WHERE code = $1 RETURNING code, user_id, expires_at`, code,
	).Scan(&lc.Code, &lc.UserId, &lc.ExpiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("link code %q: %w", code, apperrors.NotFoundError{})
	}
	if err != nil {
		return nil, fmt.Errorf("find link code: %w", err)
	}
	return &lc, nil
}
