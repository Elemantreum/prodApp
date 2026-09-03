package db

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFs embed.FS

func RunMigrations(connectionString string) error {
	source, err := iofs.New(migrationsFs, "migrations")
	if err != nil {
		return fmt.Errorf("load migrations fs: %w", err)
	}
	migrator, err := migrate.NewWithSourceInstance("iofs", source, connectionString)

	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
