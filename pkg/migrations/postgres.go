package migrations

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigratePostgres runs PostgreSQL migrations
func MigratePostgres(dbURL, migrationPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationPath),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("Successfully applied PostgreSQL migrations")
	return nil
}

// RollbackPostgres rolls back the last migration
func RollbackPostgres(dbURL, migrationPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationPath),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	log.Println("Successfully rolled back the last PostgreSQL migration")
	return nil
}
