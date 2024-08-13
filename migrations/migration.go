package migrations

import (
	"embed"
	"fmt"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/storage"

	"github.com/pressly/goose/v3"
)

//go:embed m/*.sql
var embedMigrations embed.FS

func MigrateDb(pathToDataBase string) error {
    logging.Logger.Debug(fmt.Sprintf("Migrate database at: %s", pathToDataBase))
    db, err := storage.NewDB(pathToDataBase)
    if err != nil {
        return err
    }

    goose.SetBaseFS(embedMigrations)

    if err := goose.SetDialect("sqlite3"); err != nil {
        return err
    }

    if err := goose.Up(db, "m"); err != nil {
        return err
    }
    return nil
}