package migrations

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

//go:embed *.sql
var files embed.FS

func Migrate(db *sql.DB, log *zap.SugaredLogger) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	source, err := iofs.New(files, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", source, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	log.Info("Running migrations...")

	err = migrator.Up()
	switch {
	case errors.Is(err, migrate.ErrNoChange):
		log.Info("No migrations to apply")
	case err != nil:
		return fmt.Errorf("migrate up: %w", err)
	default:
		log.Info("Migrations applied")
	}

	return nil
}
