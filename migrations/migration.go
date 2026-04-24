/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
