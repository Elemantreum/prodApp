package booking

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"prodApp/internal/platform/apperrors"
	"prodApp/internal/section"
)

func (p *PostgresBookingRepository) Waitlist(ctx context.Context, userID int) ([]int, error) {
	rows, err := p.pool.Query(ctx, `SELECT w.training_id FROM training_waitlist w JOIN trainings t ON t.id = w.training_id WHERE w.user_id = $1 AND t.section_id=$2 AND t.is_active AND t.date > now() ORDER BY t.date`, userID, section.ID(ctx))
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[int])
}

func (p *PostgresBookingRepository) WaitingMembers(ctx context.Context, id int) ([]WaitingMember, error) {
	rows, err := p.pool.Query(ctx, `SELECT u.id,u.first_name,u.last_name FROM training_waitlist w JOIN users u ON u.id=w.user_id JOIN trainings t ON t.id=w.training_id WHERE t.id=$1 AND t.section_id=$2 ORDER BY w.joined_at,u.id`, id, section.ID(ctx))
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[WaitingMember])
}

func (p *PostgresBookingRepository) History(ctx context.Context, sectionID, userID int) ([]HistoryItem, error) {
	rows, err := p.pool.Query(ctx, `SELECT b.id,b.user_id,concat_ws(' ',u.first_name,u.last_name),t.id,t.description,t.date,b.status,s.name FROM bookings b JOIN users u ON u.id=b.user_id JOIN trainings t ON t.id=b.training_id JOIN sections s ON s.id=t.section_id WHERE ($1::integer=0 OR t.section_id=$1) AND ($2::integer=0 OR b.user_id=$2) ORDER BY t.date DESC,b.id DESC`, sectionID, userID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[HistoryItem])
}

func (p *PostgresBookingRepository) LeaveWaitlist(ctx context.Context, userID, trainingID int) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM training_waitlist WHERE training_id = $1 AND user_id = $2 AND training_id IN (SELECT id FROM trainings WHERE section_id=$3)", trainingID, userID, section.ID(ctx))
	return err
}

func (p *PostgresBookingRepository) WithLockedTrainings(ctx context.Context, source, target int, apply func(RosterRepository) error) error {
	return pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		if err := lockSectionForBooking(ctx, tx); err != nil {
			return err
		}
		rows, err := tx.Query(ctx, "SELECT id FROM trainings WHERE section_id=$1 AND id IN ($2,$3) ORDER BY id FOR UPDATE", section.ID(ctx), source, target)
		if err != nil {
			return err
		}
		ids, err := pgx.CollectRows(rows, pgx.RowTo[int])
		if err != nil {
			return err
		}
		if len(ids) == 0 || (source != target && len(ids) != 2) {
			return &apperrors.NotFoundError{}
		}
		return apply(&postgresRosterRepository{tx: tx})
	})
}

type postgresRosterRepository struct {
	tx pgx.Tx
}

func (p *postgresRosterRepository) Availability(ctx context.Context, trainingID int) (free int, open, past bool, err error) {
	err = p.tx.QueryRow(ctx, `
  SELECT t.max_quantity - count(b.id), t.is_active AND t.date > now(), t.date <= now()
  FROM trainings t
  LEFT JOIN bookings b ON b.training_id=t.id
    AND (b.status='active' OR (t.date<=now() AND b.status<>'cancelled'))
  WHERE t.id=$1 GROUP BY t.id`, trainingID).Scan(&free, &open, &past)
	return
}

func (p *postgresRosterRepository) IsRegisteredMember(ctx context.Context, userID int) (bool, error) {
	var registered bool
	err := p.tx.QueryRow(ctx, `
  SELECT EXISTS(SELECT 1 FROM users u JOIN section_members m ON u.id=m.user_id
  WHERE m.section_id=$1 AND u.id=$2 AND u.is_registered)`, section.ID(ctx), userID).Scan(&registered)
	return registered, err
}

func (p *postgresRosterRepository) HasRosterEntry(ctx context.Context, userID, trainingID, exceptBookingID int) (bool, error) {
	var exists bool
	err := p.tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM bookings
		WHERE user_id=$1 AND training_id=$2 AND id<>$3 AND status<>'cancelled')`,
		userID, trainingID, exceptBookingID).Scan(&exists)
	return exists, err
}

func (p *postgresRosterRepository) WaitlistEligibility(ctx context.Context, userID, trainingID int) (eligible, booked bool, err error) {
	err = p.tx.QueryRow(ctx, `
		SELECT u.is_registered AND (u.is_admin OR EXISTS(
			SELECT 1 FROM section_members WHERE user_id=u.id AND section_id=$3
		)), EXISTS(SELECT 1 FROM bookings WHERE training_id=$2 AND user_id=u.id AND status='active')
		FROM users u WHERE u.id=$1`, userID, trainingID, section.ID(ctx)).Scan(&eligible, &booked)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, false, nil
	}
	return eligible, booked, err
}

func (p *postgresRosterRepository) JoinWaitlist(ctx context.Context, userID, trainingID int) error {
	_, err := p.tx.Exec(ctx, "INSERT INTO training_waitlist(training_id,user_id) VALUES($1,$2) ON CONFLICT DO NOTHING", trainingID, userID)
	return err
}

func (p *postgresRosterRepository) AddBooking(ctx context.Context, userID, trainingID int) error {
	return rosterWriteResult(p.tx.Exec(ctx, "INSERT INTO bookings(user_id,training_id) VALUES($1,$2)", userID, trainingID))
}

func (p *postgresRosterRepository) MoveBooking(ctx context.Context, bookingID, sourceID, targetID, userID int) error {
	return rosterWriteResult(p.tx.Exec(ctx, `
		UPDATE bookings b SET user_id=$1, training_id=$2
		FROM trainings t
		WHERE b.id=$3 AND b.training_id=$4 AND t.id=b.training_id
		  AND (b.status='active' OR ($2=$4 AND t.date<=now() AND b.status IN ('completed','not_completed')))`,
		userID, targetID, bookingID, sourceID))
}

func (p *postgresRosterRepository) DeleteBooking(ctx context.Context, bookingID, trainingID int) error {
	return rosterWriteResult(p.tx.Exec(ctx, "DELETE FROM bookings WHERE id=$1 AND training_id=$2", bookingID, trainingID))
}

func (p *postgresRosterRepository) ResetAttendance(ctx context.Context, bookingID, trainingID int) error {
	return rosterWriteResult(p.tx.Exec(ctx, "UPDATE bookings SET status='active' WHERE id=$1 AND training_id=$2 AND status IN ('completed','not_completed')", bookingID, trainingID))
}

func (p *postgresRosterRepository) AddWaitingMember(ctx context.Context, userID, trainingID int) error {
	return rosterWriteResult(p.tx.Exec(ctx, `INSERT INTO training_waitlist(training_id,user_id)
  SELECT $1,$2 WHERE NOT EXISTS(SELECT 1 FROM bookings WHERE training_id=$1 AND user_id=$2 AND status='active')`, trainingID, userID))
}

func (p *postgresRosterRepository) RemoveWaitingMember(ctx context.Context, userID, trainingID int) error {
	return rosterWriteResult(p.tx.Exec(ctx, "DELETE FROM training_waitlist WHERE training_id=$1 AND user_id=$2", trainingID, userID))
}

func rosterWriteResult(tag pgconn.CommandTag, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return &apperrors.ConflictError{Resource: "участник", Reason: "уже записан"}
	}
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return &apperrors.NotFoundError{}
	}
	return nil
}
