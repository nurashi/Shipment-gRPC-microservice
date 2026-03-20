package postgres

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
)

func RunMigrations(connString string, migrationsDir string) error {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		return fmt.Errorf("open migration connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping migration connection: %w", err)
	}

	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	current, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("get current migration version: %w", err)
	}
	log.Printf("current migration version: %d", current)

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	after, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("get migration version after apply: %w", err)
	}

	if after == current {
		log.Printf("migrations up to date at version %d", after)
	} else {
		log.Printf("migrated from version %d to %d", current, after)
	}

	return nil
}
