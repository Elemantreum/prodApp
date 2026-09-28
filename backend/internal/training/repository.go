package training

import (
	"context"
	"errors"
	"fmt"
	"prodApp/internal/platform/apperrors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTrainingRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTrainingRepository(pool *pgxpool.Pool) *PostgresTrainingRepository {
	if pool == nil {
		panic("pool is nil")
	}

	return &PostgresTrainingRepository{pool: pool}
}

func (p *PostgresTrainingRepository) Create(ctx context.Context, description string, date time.Time, maxQuantity int, duration time.Duration) (*Training, error) {
	row := p.pool.QueryRow(ctx,
		"INSERT INTO trainings (description, date, max_quantity, duration_ns) VALUES ($1, $2, $3, $4) "+
			"RETURNING id, description, date, max_quantity, duration_ns, is_active, created_at",
		description, date, maxQuantity, duration,
	)

	var tr Training
	err := row.Scan(&tr.Id, &tr.Description, &tr.Date, &tr.MaxQuantity, &tr.Duration, &tr.IsActive, &tr.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("error inserting training: %w", err)
	}

	tr.BookedCount = 0

	return &tr, nil
}

func (p *PostgresTrainingRepository) Update(ctx context.Context, id int, description string, date time.Time, maxQuantity int, duration time.Duration, isActive bool) error {
	tag, err := p.pool.Exec(ctx,
		"UPDATE trainings SET description = $2, date = $3, max_quantity = $4, duration_ns = $5, is_active = $6 WHERE id = $1",
		id, description, date, maxQuantity, duration, isActive,
	)

	if err != nil {
		return fmt.Errorf("error updating training: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("error updating training: no rows affected")
	}

	return nil
}

func (p *PostgresTrainingRepository) Delete(ctx context.Context, id int) error {
	tag, err := p.pool.Exec(ctx, "DELETE FROM trainings WHERE id = $1", id)

	if err != nil {
		return fmt.Errorf("error deleting training with id %d: %w", id, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("training with id %d not found", id)
	}

	return nil
}

func (p *PostgresTrainingRepository) DeactivatePast(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, "UPDATE trainings SET is_active = FALSE WHERE date < now() AND is_active = TRUE")

	if err != nil {
		return fmt.Errorf("error deactivating past trainings: %w", err)
	}

	return nil
}

func (p *PostgresTrainingRepository) HasBookings(ctx context.Context, trainingId int) (bool, error) {
	var exists bool
	err := p.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM bookings WHERE training_id = $1)", trainingId).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("error checking bookings for training %d: %w", trainingId, err)
	}

	return exists, nil
}

const selectTrainingsWithBookedCount = `
	SELECT t.id, t.description, t.date, t.max_quantity, t.duration_ns, t.is_active, t.created_at, t.recurring_training_id,
	       COUNT(b.id) FILTER (WHERE b.status = 'active') AS booked_count
	FROM trainings t
	LEFT JOIN bookings b ON b.training_id = t.id`

func (p *PostgresTrainingRepository) FindById(ctx context.Context, id int) (*Training, error) {
	row := p.pool.QueryRow(ctx,
		selectTrainingsWithBookedCount+" WHERE t.id = $1 GROUP BY t.id", id)

	var tr Training
	err := row.Scan(&tr.Id, &tr.Description, &tr.Date, &tr.MaxQuantity, &tr.Duration, &tr.IsActive, &tr.CreatedAt, &tr.RecurringTrainingId, &tr.BookedCount)

	if err != nil {
		return nil, fmt.Errorf("training with id %v: %w", id, err)
	}

	return &tr, nil
}

func (p *PostgresTrainingRepository) GetAll(ctx context.Context) ([]Training, error) {
	rows, err := p.pool.Query(ctx, selectTrainingsWithBookedCount+" GROUP BY t.id ORDER BY t.id DESC")

	if err != nil {
		return nil, fmt.Errorf("error getting trainings: %w", err)
	}
	defer rows.Close()

	trainings := make([]Training, 0)

	for rows.Next() {
		var tr Training

		err := rows.Scan(&tr.Id, &tr.Description, &tr.Date, &tr.MaxQuantity, &tr.Duration, &tr.IsActive, &tr.CreatedAt, &tr.RecurringTrainingId, &tr.BookedCount)

		if err != nil {
			return nil, fmt.Errorf("error scanning training: %w", err)
		}

		trainings = append(trainings, tr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trainings: %w", err)
	}

	return trainings, nil
}

func (p *PostgresTrainingRepository) CreateFromOccurrence(ctx context.Context, recurringId int, description string, date time.Time, maxQuantity int, duration time.Duration) (*Training, error) {
	row := p.pool.QueryRow(ctx,
		`INSERT INTO trainings (description, date, max_quantity, duration_ns, recurring_training_id)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (recurring_training_id, date) DO NOTHING
		 RETURNING id, description, date, max_quantity, duration_ns, is_active, created_at`,
		description, date, maxQuantity, duration, recurringId,
	)

	var tr Training
	err := row.Scan(&tr.Id, &tr.Description, &tr.Date, &tr.MaxQuantity, &tr.Duration, &tr.IsActive, &tr.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error inserting recurring occurrence: %w", err)
	}

	tr.BookedCount = 0
	return &tr, nil
}

type PostgresRecurringTrainingRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRecurringTrainingRepository(pool *pgxpool.Pool) *PostgresRecurringTrainingRepository {
	if pool == nil {
		panic("pool is nil")
	}

	return &PostgresRecurringTrainingRepository{pool: pool}
}

func (p *PostgresRecurringTrainingRepository) Create(ctx context.Context, description string, anchorDate time.Time, maxQuantity int, duration time.Duration, leadTime time.Duration) (*RecurringTraining, error) {
	row := p.pool.QueryRow(ctx,
		`INSERT INTO recurring_trainings (description, anchor_date, max_quantity, duration_ns, lead_time_ns)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, description, anchor_date, max_quantity, duration_ns, lead_time_ns, is_active, created_at`,
		description, anchorDate, maxQuantity, duration, leadTime,
	)

	var rt RecurringTraining
	err := row.Scan(&rt.Id, &rt.Description, &rt.AnchorDate, &rt.MaxQuantity, &rt.Duration, &rt.LeadTime, &rt.IsActive, &rt.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("error inserting recurring training: %w", err)
	}

	return &rt, nil
}

func (p *PostgresRecurringTrainingRepository) Update(ctx context.Context, id int, description string, anchorDate time.Time, maxQuantity int, duration time.Duration, leadTime time.Duration, isActive bool) (*RecurringTraining, error) {
	row := p.pool.QueryRow(ctx,
		`UPDATE recurring_trainings
		 SET description = $2, anchor_date = $3, max_quantity = $4, duration_ns = $5, lead_time_ns = $6, is_active = $7
		 WHERE id = $1
		 RETURNING id, description, anchor_date, max_quantity, duration_ns, lead_time_ns, is_active, created_at`,
		id, description, anchorDate, maxQuantity, duration, leadTime, isActive,
	)

	var rt RecurringTraining
	err := row.Scan(&rt.Id, &rt.Description, &rt.AnchorDate, &rt.MaxQuantity, &rt.Duration, &rt.LeadTime, &rt.IsActive, &rt.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("recurring training with id %d: %w", id, apperrors.NotFoundError{})
	}
	if err != nil {
		return nil, fmt.Errorf("error updating recurring training: %w", err)
	}

	return &rt, nil
}

func (p *PostgresRecurringTrainingRepository) GetAll(ctx context.Context) ([]RecurringTraining, error) {
	rows, err := p.pool.Query(ctx,
		"SELECT id, description, anchor_date, max_quantity, duration_ns, lead_time_ns, is_active, created_at FROM recurring_trainings ORDER BY id DESC")

	if err != nil {
		return nil, fmt.Errorf("error getting recurring trainings: %w", err)
	}
	defer rows.Close()

	rules := make([]RecurringTraining, 0)

	for rows.Next() {
		var rt RecurringTraining

		err := rows.Scan(&rt.Id, &rt.Description, &rt.AnchorDate, &rt.MaxQuantity, &rt.Duration, &rt.LeadTime, &rt.IsActive, &rt.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("error scanning recurring training: %w", err)
		}

		rules = append(rules, rt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recurring trainings: %w", err)
	}

	return rules, nil
}

func (p *PostgresRecurringTrainingRepository) FindActive(ctx context.Context) ([]RecurringTraining, error) {
	rows, err := p.pool.Query(ctx,
		"SELECT id, description, anchor_date, max_quantity, duration_ns, lead_time_ns, is_active, created_at FROM recurring_trainings WHERE is_active = TRUE ORDER BY id DESC")

	if err != nil {
		return nil, fmt.Errorf("error getting active recurring trainings: %w", err)
	}
	defer rows.Close()

	rules := make([]RecurringTraining, 0)

	for rows.Next() {
		var rt RecurringTraining

		err := rows.Scan(&rt.Id, &rt.Description, &rt.AnchorDate, &rt.MaxQuantity, &rt.Duration, &rt.LeadTime, &rt.IsActive, &rt.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("error scanning recurring training: %w", err)
		}

		rules = append(rules, rt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active recurring trainings: %w", err)
	}

	return rules, nil
}
