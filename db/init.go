package db

import (
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Init(connString string) error {
	slog.Info("Initializing database...")

	m, err := migrate.New(
		"file://db/migrations",
		connString)

	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		if err.Error() == "no change" {
			slog.Info("Db Migration Complete", "status", err)
		} else {
			return err
		}
	}

	return nil
}
