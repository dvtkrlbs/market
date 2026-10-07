package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/dvtkrlbs/market/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Migrate applies all embedded database migrations that have not run yet.
func Migrate(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open postgres for migration: %w", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping postgres for migration: %w", err)
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("open embedded migrations: %w", err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres migration driver: %w", err)
	}
	runner, err := migrate.NewWithInstance("iofs", source, "market", driver)
	if err != nil {
		return fmt.Errorf("create migration runner: %w", err)
	}
	if err := runner.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
