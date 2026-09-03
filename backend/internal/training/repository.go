package training

import (
	"context"
	"fmt"
	"time"

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

func (p *PostgresTrainingRepository) FindById(ctx context.Context, id int) (*Training, error) {
	row := p.pool.QueryRow(ctx,
		"SELECT id, description, date, max_quantity, duration_ns, is_active, created_at FROM trainings WHERE id = $1", id)

	var tr Training
	err := row.Scan(&tr.Id, &tr.Description, &tr.Date, &tr.MaxQuantity, &tr.Duration, &tr.IsActive, &tr.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("training with id %v: %w", id, err)
	}

	return &tr, nil
}

func (p *PostgresTrainingRepository) GetAll(ctx context.Context) ([]Training, error) {
	rows, err := p.pool.Query(ctx,
		"SELECT id, description, date, max_quantity, duration_ns, is_active, created_at FROM trainings")

	if err != nil {
		return nil, fmt.Errorf("error getting trainings: %w", err)
	}
	defer rows.Close()

	trainings := make([]Training, 0)

	for rows.Next() {
		var tr Training

		err := rows.Scan(&tr.Id, &tr.Description, &tr.Date, &tr.MaxQuantity, &tr.Duration, &tr.IsActive, &tr.CreatedAt)

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
