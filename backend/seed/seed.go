package seed
package seed

import (
	"database/sql"
	"fmt"
	"time"
)

func LoadDefaultData(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS windows (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP);`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS media (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, type TEXT NOT NULL, url TEXT, default_duration INTEGER NOT NULL DEFAULT 5, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP);`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS playlist_items (id INTEGER PRIMARY KEY AUTOINCREMENT, window_id INTEGER NOT NULL, media_id INTEGER NOT NULL, duration INTEGER NOT NULL, position INTEGER NOT NULL, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY(window_id) REFERENCES windows(id) ON DELETE CASCADE, FOREIGN KEY(media_id) REFERENCES media(id) ON DELETE RESTRICT);`); err != nil {
		return err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS sync_events (id INTEGER PRIMARY KEY AUTOINCREMENT, media_id INTEGER NOT NULL, start_time TEXT NOT NULL, duration INTEGER NOT NULL, end_time TEXT NOT NULL, status TEXT NOT NULL, created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY(media_id) REFERENCES media(id) ON DELETE RESTRICT);`); err != nil {
		return err
	}

	windows := []string{"Window 1", "Window 2", "Window 3"}
	for _, name := range windows {
		var exists int
		if err := db.QueryRow("SELECT COUNT(*) FROM windows WHERE name = ?", name).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			if _, err := db.Exec("INSERT INTO windows(name, created_at, updated_at) VALUES (?, ?, ?)", name, time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339)); err != nil {
				return err
			}
		}
	}

	media := []struct {
		name string
		mediaType string
		url string
		duration int
	}{
		{"M1", "image", "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80", 5},
		{"M2", "image", "https://images.unsplash.com/photo-1493246507139-91e8fad9978e?auto=format&fit=crop&w=1200&q=80", 6},
		{"M3", "video", "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4", 8},
		{"M4", "image", "https://images.unsplash.com/photo-1470770841072-f978cf4d019e?auto=format&fit=crop&w=1200&q=80", 4},
		{"M5", "image", "https://images.unsplash.com/photo-1501785888041-af3ef285b470?auto=format&fit=crop&w=1200&q=80", 7},
		{"M6", "video", "https://www.w3schools.com/html/mov_bbb.mp4", 9},
		{"Blank", "blank", "", 4},
	}
	for _, item := range media {
		var exists int
		if err := db.QueryRow("SELECT COUNT(*) FROM media WHERE name = ?", item.name).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			if _, err := db.Exec("INSERT INTO media(name, type, url, default_duration, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", item.name, item.mediaType, item.url, item.duration, time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339)); err != nil {
				return err
			}
		}
	}

	defaultPlaylists := map[string][]struct {
		name string
		position int
	}{
		"Window 1": {{"M1", 1}, {"M2", 2}, {"M3", 3}},
		"Window 2": {{"M2", 1}, {"M4", 2}, {"M5", 3}},
		"Window 3": {{"M1", 1}, {"M5", 2}, {"M6", 3}},
	}
	for windowName, items := range defaultPlaylists {
		var windowID int
		if err := db.QueryRow("SELECT id FROM windows WHERE name = ?", windowName).Scan(&windowID); err != nil {
			return err
		}
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM playlist_items WHERE window_id = ?", windowID).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		for _, item := range items {
			var mediaID int
			if err := db.QueryRow("SELECT id FROM media WHERE name = ?", item.name).Scan(&mediaID); err != nil {
				return err
			}
			var duration int
			if err := db.QueryRow("SELECT default_duration FROM media WHERE id = ?", mediaID).Scan(&duration); err != nil {
				return err
			}
			if _, err := db.Exec("INSERT INTO playlist_items(window_id, media_id, duration, position, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", windowID, mediaID, duration, item.position, time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339)); err != nil {
				return fmt.Errorf("insert seeded playlist item: %w", err)
			}
		}
	}
	return nil
}
