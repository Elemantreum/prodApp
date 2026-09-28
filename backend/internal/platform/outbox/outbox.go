package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Entry struct {
	Id      int64
	Topic   string
	Payload []byte
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func Create[T any](ctx context.Context, r *Repository, event T) error {
	topic := reflect.TypeFor[T]().Name()

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event for topic %q: %w", topic, err)
	}

	_, err = r.pool.Exec(ctx, `INSERT INTO outbox (topic, payload) VALUES ($1, $2)`, topic, payload)
	if err != nil {
		return fmt.Errorf("insert outbox entry for topic %q: %w", topic, err)
	}

	return nil
}

func (r *Repository) FindUnpublished(ctx context.Context, limit int) ([]Entry, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, topic, payload FROM outbox WHERE published_at IS NULL ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("query unpublished outbox entries: %w", err)
	}
	defer rows.Close()

	entries := make([]Entry, 0, limit)
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.Id, &e.Topic, &e.Payload); err != nil {
			return nil, fmt.Errorf("scan outbox entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox entries: %w", err)
	}

	return entries, nil
}

func (r *Repository) MarkPublished(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox SET published_at = $2 WHERE id = $1`, id, time.Now())
	if err != nil {
		return fmt.Errorf("mark outbox entry %d published: %w", id, err)
	}
	return nil
}
