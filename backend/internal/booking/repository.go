package booking

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"prodApp/internal/platform/apperrors"
	"prodApp/internal/section"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresBookingRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresBookingRepository(pool *pgxpool.Pool) *PostgresBookingRepository {
	if pool == nil {
		panic("pool is nil")
	}

	return &PostgresBookingRepository{pool: pool}
}

func (p *PostgresBookingRepository) Create(ctx context.Context, booking *Booking) (*Booking, error) {
	var saved Booking
	err := pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		if err := lockSectionForBooking(ctx, tx); err != nil {
			return err
		}
		free, err := lockAvailableTraining(ctx, tx, booking.TrainingId)
		if err != nil {
			return err
		}
		if free <= 0 {
			return &apperrors.ConflictError{Resource: "тренировка", Reason: "нет свободных мест"}
		}
		err = tx.QueryRow(ctx, `INSERT INTO bookings (user_id, training_id)
			SELECT u.id, $2 FROM users u WHERE u.id=$1 AND u.is_registered
				AND (u.is_admin OR EXISTS(SELECT 1 FROM section_members WHERE section_id=$3 AND user_id=u.id))
            RETURNING id, user_id, training_id, status, booking_date, unbooking_date`, booking.UserId, booking.TrainingId, section.ID(ctx)).Scan(&saved.Id, &saved.UserId, &saved.TrainingId, &saved.Status, &saved.BookingDate, &saved.UnbookingDate)
		if errors.Is(err, pgx.ErrNoRows) {
			return &apperrors.ForbiddenError{Reason: "нужно заполнить профиль и иметь доступ к секции"}
		}
		var conflict *pgconn.PgError
		if errors.As(err, &conflict) && conflict.Code == "23505" {
			return &apperrors.ConflictError{Resource: "тренировка", Reason: "уже есть запись"}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return &saved, nil
}

func lockSectionForBooking(ctx context.Context, tx pgx.Tx) error {
	var id int
	err := tx.QueryRow(ctx, "SELECT id FROM sections WHERE id=$1 FOR KEY SHARE", section.ID(ctx)).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return &apperrors.NotFoundError{}
	}
	return err
}

func (p *PostgresBookingRepository) Update(ctx context.Context, booking *Booking) error {
	tag, err := p.pool.Exec(ctx,
		"UPDATE bookings SET status = $4, unbooking_date = $5 WHERE id = $1 AND user_id=$2 AND training_id=$3 AND status <> 'cancelled'",
		booking.Id, booking.UserId, booking.TrainingId, booking.Status, booking.UnbookingDate,
	)

	if err != nil {
		return fmt.Errorf("error updating booking: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return &apperrors.ConflictError{Resource: "запись", Reason: "состав тренировки изменился"}
	}

	return nil
}

func (p *PostgresBookingRepository) Delete(ctx context.Context, id int) error {
	tag, err := p.pool.Exec(ctx, "DELETE FROM bookings WHERE id = $1 AND training_id IN (SELECT id FROM trainings WHERE section_id=$2)", id, section.ID(ctx))

	if err != nil {
		return fmt.Errorf("error deleting booking with id %d: %w", id, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("booking with id %d not found", id)
	}

	return nil
}

func (p *PostgresBookingRepository) FindById(ctx context.Context, id int) (*Booking, error) {
	row := p.pool.QueryRow(ctx,
		"SELECT id, user_id, training_id, status, booking_date, unbooking_date FROM bookings WHERE id = $1 AND training_id IN (SELECT id FROM trainings WHERE section_id=$2)", id, section.ID(ctx))

	var booking Booking
	err := row.Scan(&booking.Id, &booking.UserId, &booking.TrainingId, &booking.Status, &booking.BookingDate, &booking.UnbookingDate)

	if err != nil {
		return nil, fmt.Errorf("booking with id %v: %w", id, err)
	}

	return &booking, nil
}

func (p *PostgresBookingRepository) FindByUserId(ctx context.Context, userId int) ([]Booking, error) {
	rows, err := p.pool.Query(ctx,
		"SELECT id, user_id, training_id, status, booking_date, unbooking_date FROM bookings WHERE user_id = $1 AND training_id IN (SELECT id FROM trainings WHERE section_id=$2) ORDER BY id DESC", userId, section.ID(ctx))

	if err != nil {
		return nil, fmt.Errorf("error getting bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]Booking, 0)

	for rows.Next() {
		var booking Booking

		err := rows.Scan(&booking.Id, &booking.UserId, &booking.TrainingId, &booking.Status, &booking.BookingDate, &booking.UnbookingDate)

		if err != nil {
			return nil, fmt.Errorf("error scanning bookings: %w", err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bookings: %w", err)
	}

	return bookings, nil
}

func (p *PostgresBookingRepository) FindByTrainingId(ctx context.Context, trainingId int) ([]Booking, error) {
	rows, err := p.pool.Query(ctx,
		"SELECT id, user_id, training_id, status, booking_date, unbooking_date FROM bookings WHERE training_id = $1 AND status = 'active'", trainingId)

	if err != nil {
		return nil, fmt.Errorf("error getting bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]Booking, 0)

	for rows.Next() {
		var booking Booking

		err := rows.Scan(&booking.Id, &booking.UserId, &booking.TrainingId, &booking.Status, &booking.BookingDate, &booking.UnbookingDate)

		if err != nil {
			return nil, fmt.Errorf("error scanning bookings: %w", err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bookings: %w", err)
	}

	return bookings, nil
}

func (p *PostgresBookingRepository) FindAllByTrainingId(ctx context.Context, trainingId int) ([]Booking, error) {
	rows, err := p.pool.Query(ctx,
		"SELECT id, user_id, training_id, status, booking_date, unbooking_date FROM bookings WHERE training_id = $1 ORDER BY booking_date DESC", trainingId)

	if err != nil {
		return nil, fmt.Errorf("error getting bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]Booking, 0)

	for rows.Next() {
		var booking Booking

		err := rows.Scan(&booking.Id, &booking.UserId, &booking.TrainingId, &booking.Status, &booking.BookingDate, &booking.UnbookingDate)

		if err != nil {
			return nil, fmt.Errorf("error scanning bookings: %w", err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bookings: %w", err)
	}

	return bookings, nil
}

func (p *PostgresBookingRepository) GetAll(ctx context.Context) ([]Booking, error) {
	rows, err := p.pool.Query(ctx,
		"SELECT id, user_id, training_id, status, booking_date, unbooking_date FROM bookings")

	if err != nil {
		return nil, fmt.Errorf("error getting bookings: %w", err)
	}
	defer rows.Close()

	bookings := make([]Booking, 0)

	for rows.Next() {
		var booking Booking

		err := rows.Scan(&booking.Id, &booking.UserId, &booking.TrainingId, &booking.Status, &booking.BookingDate, &booking.UnbookingDate)

		if err != nil {
			return nil, fmt.Errorf("error scanning bookings: %w", err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bookings: %w", err)
	}

	return bookings, nil
}

func (p *PostgresBookingRepository) GetAntiLeaderboard(ctx context.Context) ([]SinnerStats, error) {
	rows, err := p.pool.Query(ctx, `
		WITH stats AS (
			SELECT
				u.id AS user_id,
				u.first_name,
				u.last_name,
				COUNT(*) FILTER (
					WHERE b.status = 'cancelled'
					  AND b.unbooking_date IS NOT NULL
					  AND b.unbooking_date < t.date
					  AND t.date - b.unbooking_date < INTERVAL '2 hours'
					  AND NOT EXISTS (
						SELECT 1 FROM bookings b2
						WHERE b2.user_id = b.user_id
						  AND b2.training_id = b.training_id
						  AND b2.booking_date > b.unbooking_date
					  )
				) AS late_cancellations,
				COUNT(*) FILTER (WHERE b.status = 'not_completed') AS no_shows
			FROM bookings b
			JOIN trainings t ON t.id = b.training_id
			JOIN users u ON u.id = b.user_id
			WHERE t.section_id=$1
 GROUP BY u.id, u.first_name, u.last_name
		)
		SELECT user_id, first_name, last_name, late_cancellations, no_shows,
		       late_cancellations * 1 + no_shows * 2 AS penalty_points
		FROM stats
		WHERE late_cancellations > 0 OR no_shows > 0
		ORDER BY penalty_points DESC, no_shows DESC`, section.ID(ctx))
	if err != nil {
		return nil, fmt.Errorf("error querying anti leaderboard: %w", err)
	}
	defer rows.Close()

	sinners := make([]SinnerStats, 0)
	for rows.Next() {
		var s SinnerStats
		if err := rows.Scan(&s.UserId, &s.FirstName, &s.LastName, &s.LateCancellations, &s.NoShows, &s.PenaltyPoints); err != nil {
			return nil, fmt.Errorf("error scanning anti leaderboard row: %w", err)
		}
		sinners = append(sinners, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating anti leaderboard: %w", err)
	}

	return sinners, nil
}

func lockAvailableTraining(ctx context.Context, tx pgx.Tx, trainingID int) (int, error) {
	var capacity int
	var open bool
	err := tx.QueryRow(ctx, "SELECT max_quantity, is_active AND date > now() FROM trainings WHERE id = $1 AND section_id=$2 FOR UPDATE", trainingID, section.ID(ctx)).Scan(&capacity, &open)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, &apperrors.NotFoundError{}
	}
	if err != nil {
		return 0, err
	}
	if !open {
		return 0, &apperrors.ConflictError{Resource: "тренировка", Reason: "запись закрыта"}
	}
	var used int
	err = tx.QueryRow(ctx, "SELECT count(*) FROM bookings WHERE training_id = $1 AND status = 'active'", trainingID).Scan(&used)
	return capacity - used, err
}
