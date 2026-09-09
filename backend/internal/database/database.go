package database
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func Initialize(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "./data/media_scheduler.db"
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	if err := createSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	return db, nil
}

func createSchema(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS windows (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS media (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('image','video','blank')),
			url TEXT,
			default_duration INTEGER NOT NULL DEFAULT 5,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS playlist_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			window_id INTEGER NOT NULL,
			media_id INTEGER NOT NULL,
			duration INTEGER NOT NULL,
			position INTEGER NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(window_id) REFERENCES windows(id) ON DELETE CASCADE,
			FOREIGN KEY(media_id) REFERENCES media(id) ON DELETE RESTRICT
		);`,
		`CREATE TABLE IF NOT EXISTS sync_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			media_id INTEGER NOT NULL,
			start_time TEXT NOT NULL,
			duration INTEGER NOT NULL,
			end_time TEXT NOT NULL,
			status TEXT NOT NULL CHECK(status IN ('active','expired','stopped')),
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(media_id) REFERENCES media(id) ON DELETE RESTRICT
		);`,
		`CREATE INDEX IF NOT EXISTS idx_playlist_window_position ON playlist_items(window_id, position);`,
		`CREATE INDEX IF NOT EXISTS idx_sync_status_start ON sync_events(status, start_time);`,
	}

	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}
