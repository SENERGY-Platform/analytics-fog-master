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

package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	maxOpenConns    = 25
	maxIdleConns    = 25
	connMaxLifetime = 5 * time.Minute
)

func NewDB(pathToDatabase string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(pathToDatabase), 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dsn(pathToDatabase))
	if err != nil {
		return nil, fmt.Errorf("open sqlite3 db: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite3 db: %w", err)
	}

	configurePool(db)

	return db, nil
}

func dsn(path string) string {
	return fmt.Sprintf("file:%s?_foreign_keys=on&mode=rwc", path)
}

func configurePool(db *sql.DB) {
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
}
