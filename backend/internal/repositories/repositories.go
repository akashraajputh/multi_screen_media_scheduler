package repositories
package repositories

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"multi-screen-media-scheduler/internal/models"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWindow(name string) (*models.Window, error) {
	res, err := r.db.Exec("INSERT INTO windows(name, created_at, updated_at) VALUES (?, ?, ?)", name, time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetWindow(int(id))
}

func (r *Repository) GetWindows() ([]models.Window, error) {
	rows, err := r.db.Query("SELECT id, name, created_at, updated_at FROM windows ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	windows := make([]models.Window, 0)
	for rows.Next() {
		var w models.Window
		if err := rows.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		windows = append(windows, w)
	}
	return windows, rows.Err()
}

func (r *Repository) GetWindow(id int) (*models.Window, error) {
	row := r.db.QueryRow("SELECT id, name, created_at, updated_at FROM windows WHERE id = ?", id)
	var w models.Window
	if err := row.Scan(&w.ID, &w.Name, &w.CreatedAt, &w.UpdatedAt); err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *Repository) DeleteWindow(id int) error {
	_, err := r.db.Exec("DELETE FROM windows WHERE id = ?", id)
	return err
}

func (r *Repository) GetMedia() ([]models.Media, error) {
	rows, err := r.db.Query("SELECT id, name, type, url, default_duration, created_at, updated_at FROM media ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.Media, 0)
	for rows.Next() {
		var m models.Media
		if err := rows.Scan(&m.ID, &m.Name, &m.Type, &m.URL, &m.DefaultDuration, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (r *Repository) CreateMedia(name, mediaType, url string, defaultDuration int) (*models.Media, error) {
	res, err := r.db.Exec("INSERT INTO media(name, type, url, default_duration, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", name, mediaType, url, defaultDuration, time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetMediaByID(int(id))
}

func (r *Repository) GetMediaByID(id int) (*models.Media, error) {
	row := r.db.QueryRow("SELECT id, name, type, url, default_duration, created_at, updated_at FROM media WHERE id = ?", id)
	var m models.Media
	if err := row.Scan(&m.ID, &m.Name, &m.Type, &m.URL, &m.DefaultDuration, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) AddPlaylistItem(windowID, mediaID, duration, position int) (*models.PlaylistItem, error) {
	if position <= 0 {
		position = 1
	}
	res, err := r.db.Exec("INSERT INTO playlist_items(window_id, media_id, duration, position, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)", windowID, mediaID, duration, position, time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetPlaylistItemByID(int(id))
}

func (r *Repository) GetPlaylistItems(windowID int) ([]models.PlaylistItem, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.window_id, p.media_id, m.type, m.url, p.duration, p.position, p.created_at, p.updated_at
		FROM playlist_items p
		JOIN media m ON m.id = p.media_id
		WHERE p.window_id = ?
		ORDER BY p.position ASC, p.id ASC`, windowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.PlaylistItem, 0)
	for rows.Next() {
		var item models.PlaylistItem
		if err := rows.Scan(&item.ID, &item.WindowID, &item.MediaID, &item.MediaType, &item.MediaURL, &item.Duration, &item.Position, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Media = &models.Media{ID: item.MediaID, Type: item.MediaType, URL: item.MediaURL}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetPlaylistItemByID(itemID int) (*models.PlaylistItem, error) {
	row := r.db.QueryRow(`
		SELECT p.id, p.window_id, p.media_id, m.type, m.url, p.duration, p.position, p.created_at, p.updated_at
		FROM playlist_items p
		JOIN media m ON m.id = p.media_id
		WHERE p.id = ?`, itemID)
	var item models.PlaylistItem
	if err := row.Scan(&item.ID, &item.WindowID, &item.MediaID, &item.MediaType, &item.MediaURL, &item.Duration, &item.Position, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	item.Media = &models.Media{ID: item.MediaID, Type: item.MediaType, URL: item.MediaURL}
	return &item, nil
}

func (r *Repository) UpdatePlaylistItem(itemID, duration, position int) (*models.PlaylistItem, error) {
	if position <= 0 {
		position = 1
	}
	_, err := r.db.Exec("UPDATE playlist_items SET duration = ?, position = ?, updated_at = ? WHERE id = ?", duration, position, time.Now().UTC().Format(time.RFC3339), itemID)
	if err != nil {
		return nil, err
	}
	return r.GetPlaylistItemByID(itemID)
}

func (r *Repository) DeletePlaylistItem(itemID int) error {
	_, err := r.db.Exec("DELETE FROM playlist_items WHERE id = ?", itemID)
	return err
}

func (r *Repository) ReorderPlaylist(windowID int, orderedIDs []int) error {
	if len(orderedIDs) == 0 {
		return nil
	}
	for index, itemID := range orderedIDs {
		if _, err := r.db.Exec("UPDATE playlist_items SET position = ?, updated_at = ? WHERE id = ? AND window_id = ?", index+1, time.Now().UTC().Format(time.RFC3339), itemID, windowID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) CreateSyncEvent(mediaID, duration int, start time.Time) (*models.SyncEvent, error) {
	startTime := start.UTC().Format(time.RFC3339)
	endTime := start.Add(time.Duration(duration) * time.Second).UTC().Format(time.RFC3339)
	if _, err := r.db.Exec("UPDATE sync_events SET status = 'expired' WHERE status = 'active'"); err != nil {
		return nil, err
	}
	res, err := r.db.Exec("INSERT INTO sync_events(media_id, start_time, duration, end_time, status, created_at) VALUES (?, ?, ?, ?, ?, ?)", mediaID, startTime, duration, endTime, "active", time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetActiveSyncEvent(int(id))
}

func (r *Repository) GetActiveSyncEvent(id int) (*models.SyncEvent, error) {
	row := r.db.QueryRow(`
		SELECT s.id, s.media_id, m.type, m.url, s.start_time, s.duration, s.end_time, s.status, s.created_at
		FROM sync_events s
		JOIN media m ON m.id = s.media_id
		WHERE s.id = ?`, id)
	var event models.SyncEvent
	if err := row.Scan(&event.ID, &event.MediaID, &event.MediaType, &event.MediaURL, &event.StartTime, &event.Duration, &event.EndTime, &event.Status, &event.CreatedAt); err != nil {
		return nil, err
	}
	event.Media = &models.Media{ID: event.MediaID, Type: event.MediaType, URL: event.MediaURL}
	return &event, nil
}

func (r *Repository) GetCurrentSyncEvent() (*models.SyncEvent, error) {
	row := r.db.QueryRow(`
		SELECT s.id, s.media_id, m.type, m.url, s.start_time, s.duration, s.end_time, s.status, s.created_at
		FROM sync_events s
		JOIN media m ON m.id = s.media_id
		WHERE s.status = 'active'
		ORDER BY s.created_at DESC
		LIMIT 1`)
	var event models.SyncEvent
	var start string
	var end string
	if err := row.Scan(&event.ID, &event.MediaID, &event.MediaType, &event.MediaURL, &start, &event.Duration, &end, &event.Status, &event.CreatedAt); err != nil {
		return nil, err
	}
	event.StartTime, _ = time.Parse(time.RFC3339, start)
	event.EndTime, _ = time.Parse(time.RFC3339, end)
	event.Media = &models.Media{ID: event.MediaID, Type: event.MediaType, URL: event.MediaURL}
	return &event, nil
}

func (r *Repository) StopSyncEvent() error {
	_, err := r.db.Exec("UPDATE sync_events SET status = 'stopped' WHERE status = 'active'")
	return err
}

func (r *Repository) ExpireSyncEvents() error {
	_, err := r.db.Exec("UPDATE sync_events SET status = 'expired' WHERE status = 'active' AND end_time <= ?", time.Now().UTC().Format(time.RFC3339))
	return err
}

func (r *Repository) GetMediaByName(name string) (*models.Media, error) {
	row := r.db.QueryRow("SELECT id, name, type, url, default_duration, created_at, updated_at FROM media WHERE name = ? LIMIT 1", name)
	var m models.Media
	if err := row.Scan(&m.ID, &m.Name, &m.Type, &m.URL, &m.DefaultDuration, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Repository) WindowPlaylist(windowID int) ([]models.PlaylistItem, error) {
	items, err := r.GetPlaylistItems(windowID)
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Position < items[j].Position })
	return items, nil
}

func (r *Repository) WindowCycleDuration(windowID int) (int, error) {
	items, err := r.WindowPlaylist(windowID)
	if err != nil {
		return 0, err
	}
	seconds := 0
	for _, item := range items {
		seconds += item.Duration
	}
	return seconds, nil
}

func (r *Repository) CheckMediaExists(mediaID int) bool {
	var count int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM media WHERE id = ?", mediaID).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func (r *Repository) UpsertWindow(windowID int, name string) (*models.Window, error) {
	if windowID == 0 {
		return r.CreateWindow(name)
	}
	_, err := r.db.Exec("UPDATE windows SET name = ?, updated_at = ? WHERE id = ?", name, time.Now().UTC().Format(time.RFC3339), windowID)
	if err != nil {
		return nil, err
	}
	return r.GetWindow(windowID)
}

func (r *Repository) CreateDefaultMedia() error {
	seed := []struct {
		name string
		mediaType string
		url string
		duration int
	}{
		{"M1", "image", "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80", 5},
		{"M2", "image", "https://images.unsplash.com/photo-1493246507139-91e8fad9978e?auto=format&fit=crop&w=1200&q=80", 6},
		{"M3", "video", "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4", 8},
		{"M4", "blank", "", 4},
		{"M5", "image", "https://images.unsplash.com/photo-1470770841072-f978cf4d019e?auto=format&fit=crop&w=1200&q=80", 7},
		{"M6", "video", "https://www.w3schools.com/html/mov_bbb.mp4", 9},
	}
	for _, item := range seed {
		if _, err := r.GetMediaByName(item.name); err == nil {
			continue
		}
		if _, err := r.CreateMedia(item.name, item.mediaType, item.url, item.duration); err != nil {
			return fmt.Errorf("create default media %s: %w", item.name, err)
		}
	}
	return nil
}
