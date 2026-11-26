package schedule

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type PostgresChangeRepository struct {
	dsn string
}

func NewPostgresChangeRepository(dsn string) *PostgresChangeRepository {
	return &PostgresChangeRepository{dsn: dsn}
}

func (p *PostgresChangeRepository) Listen(ctx context.Context, notify func()) error {
	conn, err := pgx.Connect(ctx, p.dsn)
	if err != nil {
		return err
	}
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
		defer cancel()
		_ = conn.Close(closeCtx)
	}()
	if _, err := conn.Exec(ctx, "LISTEN schedule_changed"); err != nil {
		return err
	}
	notify()
	for {
		if _, err := conn.WaitForNotification(ctx); err != nil {
			return err
		}
		notify()
	}
}
