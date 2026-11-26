package sectioninvite

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"prodApp/internal/platform/apperrors"
)

type PostgresInvitationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresInvitationRepository(pool *pgxpool.Pool) *PostgresInvitationRepository {
	return &PostgresInvitationRepository{pool: pool}
}

func (p *PostgresInvitationRepository) Create(ctx context.Context, sectionID, createdBy int, tokenHash string, expiresAt time.Time, maxUses int) (Invitation, error) {
	invite, err := scanInvitation(p.pool.QueryRow(ctx, `
		INSERT INTO section_invites (section_id, created_by, token_hash, expires_at, max_uses)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, expires_at, max_uses, uses, revoked_at`, sectionID, createdBy, tokenHash, expiresAt, maxUses))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return Invitation{}, &apperrors.NotFoundError{}
	}
	return invite, err
}

func (p *PostgresInvitationRepository) List(ctx context.Context, sectionID int) ([]Invitation, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, expires_at, max_uses, uses, revoked_at
		FROM section_invites WHERE section_id=$1 ORDER BY id DESC`, sectionID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Invitation, error) {
		return scanInvitation(row)
	})
}

func (p *PostgresInvitationRepository) Revoke(ctx context.Context, sectionID, inviteID int) error {
	tag, err := p.pool.Exec(ctx, `
		UPDATE section_invites SET revoked_at=COALESCE(revoked_at, clock_timestamp())
		WHERE section_id=$1 AND id=$2`, sectionID, inviteID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &apperrors.NotFoundError{}
	}
	return nil
}

const selectInviteState = `
	SELECT i.id, i.expires_at, i.max_uses, i.uses, i.revoked_at, s.id, s.name, s.logo_url,
	       EXISTS (SELECT 1 FROM section_members m WHERE m.section_id=s.id AND m.user_id=$2)
	FROM section_invites i JOIN sections s ON s.id=i.section_id
	WHERE i.token_hash=$1`

func (p *PostgresInvitationRepository) Lookup(ctx context.Context, tokenHash string, userID int) (InviteState, error) {
	return scanInviteState(p.pool.QueryRow(ctx, selectInviteState, tokenHash, userID))
}

func (p *PostgresInvitationRepository) WithLockedInvite(ctx context.Context, tokenHash string, userID int, apply func(InviteState, bool, InviteWriter) error) error {
	return pgx.BeginTxFunc(ctx, p.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		var sectionID int
		err := tx.QueryRow(ctx, `
			SELECT id FROM sections
			WHERE id=(SELECT section_id FROM section_invites WHERE token_hash=$1)
			FOR NO KEY UPDATE`, tokenHash).Scan(&sectionID)
		if err != nil {
			return inviteRowError(err)
		}
		var registered bool
		err = tx.QueryRow(ctx, `SELECT is_registered FROM users WHERE id=$1 FOR NO KEY UPDATE`, userID).Scan(&registered)
		if err != nil {
			return inviteRowError(err)
		}
		state, err := scanInviteState(tx.QueryRow(ctx, selectInviteState+" AND s.id=$3 FOR NO KEY UPDATE OF i", tokenHash, userID, sectionID))
		if err != nil {
			return err
		}
		return apply(state, registered, &postgresInviteWriter{tx: tx, sectionID: sectionID, userID: userID, inviteID: state.Invitation.ID})
	})
}

type postgresInviteWriter struct {
	tx        pgx.Tx
	sectionID int
	userID    int
	inviteID  int
}

func (p *postgresInviteWriter) InsertMember(ctx context.Context) (bool, error) {
	tag, err := p.tx.Exec(ctx, `
		INSERT INTO section_members (section_id, user_id, is_admin) VALUES ($1, $2, false)
		ON CONFLICT (section_id, user_id) DO NOTHING`, p.sectionID, p.userID)
	return tag.RowsAffected() == 1, err
}

func (p *postgresInviteWriter) Consume(ctx context.Context) error {
	_, err := p.tx.Exec(ctx, `UPDATE section_invites SET uses=uses+1 WHERE id=$1`, p.inviteID)
	return err
}

func scanInvitation(row pgx.Row) (Invitation, error) {
	var i Invitation
	err := row.Scan(&i.ID, &i.ExpiresAt, &i.MaxUses, &i.Uses, &i.RevokedAt)
	return i, err
}

func scanInviteState(row pgx.Row) (InviteState, error) {
	var state InviteState
	i, s := &state.Invitation, &state.Section
	err := row.Scan(&i.ID, &i.ExpiresAt, &i.MaxUses, &i.Uses, &i.RevokedAt, &s.ID, &s.Name, &s.Logo, &state.Member)
	return state, inviteRowError(err)
}

func inviteRowError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return &apperrors.NotFoundError{}
	}
	return err
}
