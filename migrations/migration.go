package migrations

import (
	"database/sql"
	"embed"

	"github.com/pressly/goose/v3"
)

//go:embed m/*.sql
var embedMigrations embed.FS

func MigrateDb(db *sql.DB) error {
    goose.SetBaseFS(embedMigrations)

    if err := goose.SetDialect("sqlite3"); err != nil {
        return err
    }

    if err := goose.Up(db, "m"); err != nil {
        return err
    }
    return nil
}