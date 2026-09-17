package booking

import (
	"context"
	"fmt"

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
	row := p.pool.QueryRow(ctx,
		"INSERT INTO bookings (user_id, training_id) VALUES ($1, $2) "+
			"RETURNING id, user_id, training_id, status, booking_date, unbooking_date",
		booking.UserId, booking.TrainingId,
	)

	var saved Booking
	err := row.Scan(&saved.Id, &saved.UserId, &saved.TrainingId, &saved.Status, &saved.BookingDate, &saved.UnbookingDate)

	if err != nil {
		return nil, fmt.Errorf("error inserting booking: %w", err)
	}

	return &saved, nil
}

func (p *PostgresBookingRepository) Update(ctx context.Context, booking *Booking) error {
	tag, err := p.pool.Exec(ctx,
		"UPDATE bookings SET user_id = $2, training_id = $3, status = $4, unbooking_date = $5 WHERE id = $1",
		booking.Id, booking.UserId, booking.TrainingId, booking.Status, booking.UnbookingDate,
	)

	if err != nil {
		return fmt.Errorf("error updating booking: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("error updating booking: no rows affected")
	}

	return nil
}

func (p *PostgresBookingRepository) Delete(ctx context.Context, id int) error {
	tag, err := p.pool.Exec(ctx, "DELETE FROM bookings WHERE id = $1", id)

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
		"SELECT id, user_id, training_id, status, booking_date, unbooking_date FROM bookings WHERE id = $1", id)

	var booking Booking
	err := row.Scan(&booking.Id, &booking.UserId, &booking.TrainingId, &booking.Status, &booking.BookingDate, &booking.UnbookingDate)

	if err != nil {
		return nil, fmt.Errorf("booking with id %v: %w", id, err)
	}

	return &booking, nil
}

func (p *PostgresBookingRepository) FindByUserId(ctx context.Context, userId int) ([]Booking, error) {
	rows, err := p.pool.Query(ctx,
		"SELECT id, user_id, training_id, status, booking_date, unbooking_date FROM bookings WHERE user_id = $1 ORDER BY id DESC", userId)

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

// FindAllByTrainingId — то же самое, что FindByTrainingId, но без фильтра по
// статусу. Нужно только для админской ручки "все записи на тренировку"
// (включая отменённые) — обычным пользователям такое видеть незачем, для
// них есть GetParticipants/FindByTrainingId.
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
