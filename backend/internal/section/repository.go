package section

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"prodApp/internal/platform/apperrors"
)

type PostgresSectionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSectionRepository(pool *pgxpool.Pool) *PostgresSectionRepository {
	return &PostgresSectionRepository{pool: pool}
}

func (p *PostgresSectionRepository) FindAccess(ctx context.Context, userID, sectionID int) (admin, allowed bool, err error) {
	if sectionID == 0 {
		err = p.pool.QueryRow(ctx, "SELECT is_admin, true FROM users WHERE id=$1", userID).Scan(&admin, &allowed)
	} else {
		err = p.pool.QueryRow(ctx, `
			SELECT u.is_admin OR coalesce(m.is_admin,false), u.is_admin OR m.user_id IS NOT NULL
			FROM users u CROSS JOIN sections s
			LEFT JOIN section_members m ON m.section_id=s.id AND m.user_id=u.id
			WHERE u.id=$1 AND s.id=$2`, userID, sectionID).Scan(&admin, &allowed)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return false, false, nil
	}
	return admin, allowed, err
}

func (p *PostgresSectionRepository) List(ctx context.Context, userID int) ([]SectionWithVenues, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT s.id, s.name, s.logo_url, u.is_admin OR coalesce(m.is_admin,false),
		       ARRAY(SELECT v.name FROM section_venues v WHERE v.section_id=s.id ORDER BY v.name)
		FROM users u
		CROSS JOIN sections s
		LEFT JOIN section_members m ON m.section_id=s.id AND m.user_id=u.id
		WHERE u.id=$1 AND (u.is_admin OR m.user_id IS NOT NULL)
		ORDER BY s.id`, userID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[SectionWithVenues])
}

func (p *PostgresSectionRepository) Create(ctx context.Context, input Section) (Section, error) {
	err := p.pool.QueryRow(ctx, "INSERT INTO sections(name,logo_url) VALUES($1,$2) RETURNING id", input.Name, input.Logo).Scan(&input.ID)
	return input, err
}

func (p *PostgresSectionRepository) WithLockedSection(ctx context.Context, id int, apply func(LockedSectionRepository) error) error {
	return pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		var locked int
		if err := tx.QueryRow(ctx, "SELECT id FROM sections WHERE id=$1 FOR UPDATE", id).Scan(&locked); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return &apperrors.NotFoundError{}
			}
			return err
		}
		return apply(&postgresLockedSectionRepository{tx: tx, sectionID: id})
	})
}

type postgresLockedSectionRepository struct {
	tx        pgx.Tx
	sectionID int
}

func (p *postgresLockedSectionRepository) UpdateDetails(ctx context.Context, name, logo string) error {
	_, err := p.tx.Exec(ctx, "UPDATE sections SET name=$2, logo_url=$3 WHERE id=$1", p.sectionID, name, logo)
	return err
}

func (p *postgresLockedSectionRepository) IsAdmin(ctx context.Context, userID int) (bool, error) {
	var admin bool
	err := p.tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM section_members WHERE section_id=$1 AND user_id=$2 AND is_admin)", p.sectionID, userID).Scan(&admin)
	return admin, err
}

func (p *postgresLockedSectionRepository) PutMember(ctx context.Context, userID int, admin bool) error {
	_, err := p.tx.Exec(ctx, `
		INSERT INTO section_members(section_id,user_id,is_admin) VALUES($1,$2,$3)
		ON CONFLICT(section_id,user_id) DO UPDATE SET is_admin=excluded.is_admin`, p.sectionID, userID, admin)
	return err
}

func (p *postgresLockedSectionRepository) RemoveWaitingEntries(ctx context.Context, userID int) error {
	_, err := p.tx.Exec(ctx, "DELETE FROM training_waitlist w USING trainings t WHERE w.training_id=t.id AND t.section_id=$1 AND w.user_id=$2", p.sectionID, userID)
	return err
}

func (p *postgresLockedSectionRepository) CancelUpcomingBookings(ctx context.Context, userID int) error {
	_, err := p.tx.Exec(ctx, "UPDATE bookings b SET status='cancelled',unbooking_date=now() FROM trainings t WHERE b.training_id=t.id AND t.section_id=$1 AND t.date>now() AND b.user_id=$2 AND b.status='active'", p.sectionID, userID)
	return err
}

func (p *postgresLockedSectionRepository) RemoveMember(ctx context.Context, userID int) error {
	_, err := p.tx.Exec(ctx, "DELETE FROM section_members WHERE section_id=$1 AND user_id=$2", p.sectionID, userID)
	return err
}

func (p *PostgresSectionRepository) Roster(ctx context.Context, id int) ([]Member, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT u.id,u.first_name,u.last_name,m.is_admin
		FROM users u JOIN section_members m ON u.id=m.user_id
		WHERE m.section_id=$1 ORDER BY u.last_name,u.first_name,u.id`, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Member])
}

func (p *PostgresSectionRepository) Candidates(ctx context.Context, id int) ([]Member, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT u.id,u.first_name,u.last_name,coalesce(m.is_admin,false)
		FROM users u LEFT JOIN section_members m ON m.user_id=u.id AND m.section_id=$1
		WHERE u.is_registered ORDER BY u.last_name,u.first_name,u.id`, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Member])
}

func (p *PostgresSectionRepository) DeleteEmpty(ctx context.Context, id int) (bool, error) {
	tag, err := p.pool.Exec(ctx, `DELETE FROM sections WHERE id=$1
		AND NOT EXISTS(SELECT 1 FROM trainings WHERE section_id=$1)
		AND NOT EXISTS(SELECT 1 FROM recurring_trainings WHERE section_id=$1)`, id)
	return tag.RowsAffected() != 0, err
}

func (p *PostgresSectionRepository) AddVenue(ctx context.Context, sectionID int, name string) (bool, error) {
	tag, err := p.pool.Exec(ctx, "INSERT INTO section_venues (section_id, name) VALUES ($1,$2) ON CONFLICT DO NOTHING", sectionID, name)
	return tag.RowsAffected() != 0, err
}

func (p *PostgresSectionRepository) DeleteVenue(ctx context.Context, sectionID int, name string) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM section_venues WHERE section_id=$1 AND name=$2", sectionID, name)
	return err
}
