package migrations

import (
    "embed"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/storage"

    "github.com/pressly/goose/v3"
)

//go:embed m/*.sql
var embedMigrations embed.FS

func MigrateDb(pathToDataBase string) {
    db, err := storage.NewDB(pathToDataBase)
    if err != nil {
        return
    }
    // setup database

    goose.SetBaseFS(embedMigrations)

    if err := goose.SetDialect("sqlite3"); err != nil {
        panic(err)
    }

    if err := goose.Up(db, "m"); err != nil {
        panic(err)
    }

}