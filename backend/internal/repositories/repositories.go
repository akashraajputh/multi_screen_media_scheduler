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
		var window models.Window
		if err := rows.Scan(&window.ID, &window.Name, &window.CreatedAt, &window.UpdatedAt); err != nil {
			return nil, err
		}
		windows = append(windows, window)
	}
	return windows, rows.Err()
}

func (r *Repository) GetWindow(id int) (*models.Window, error) {
	row := r.db.QueryRow("SELECT id, name, created_at, updated_at FROM windows WHERE id = ?", id)
	var window models.Window
	if err := row.Scan(&window.ID, &window.Name, &window.CreatedAt, &window.UpdatedAt); err != nil {
		return nil, err
	}
	return &window, nil
}

func (r *Repository) GetWindowByName(name string) (*models.Window, error) {
	row := r.db.QueryRow("SELECT id, name, created_at, updated_at FROM windows WHERE name = ? LIMIT 1", name)
	var window models.Window
	if err := row.Scan(&window.ID, &window.Name, &window.CreatedAt, &window.UpdatedAt); err != nil {
		return nil, err
	}
	return &window, nil
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
		var media models.Media
		if err := rows.Scan(&media.ID, &media.Name, &media.Type, &media.URL, &media.DefaultDuration, &media.CreatedAt, &media.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, media)
	}
	return items, rows.Err()
}

func (r *Repository) GetMediaByID(id int) (*models.Media, error) {
	row := r.db.QueryRow("SELECT id, name, type, url, default_duration, created_at, updated_at FROM media WHERE id = ?", id)
	var media models.Media
	if err := row.Scan(&media.ID, &media.Name, &media.Type, &media.URL, &media.DefaultDuration, &media.CreatedAt, &media.UpdatedAt); err != nil {
		return nil, err
	}
	return &media, nil
}

func (r *Repository) GetMediaByName(name string) (*models.Media, error) {
	row := r.db.QueryRow("SELECT id, name, type, url, default_duration, created_at, updated_at FROM media WHERE name = ? LIMIT 1", name)
	var media models.Media
	if err := row.Scan(&media.ID, &media.Name, &media.Type, &media.URL, &media.DefaultDuration, &media.CreatedAt, &media.UpdatedAt); err != nil {
		return nil, err
	}
	return &media, nil
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

func (r *Repository) AddPlaylistItem(windowID, mediaID, duration, position int) (*models.PlaylistItem, error) {
	if position <= 0 {
		var maxPosition int
		if err := r.db.QueryRow("SELECT COALESCE(MAX(position), 0) FROM playlist_items WHERE window_id = ?", windowID).Scan(&maxPosition); err != nil {
			return nil, err
		}
		position = maxPosition + 1
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
		SELECT p.id, p.window_id, p.media_id, m.name, m.type, m.url, p.duration, p.position, p.created_at, p.updated_at
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
		var mediaName string
		if err := rows.Scan(&item.ID, &item.WindowID, &item.MediaID, &mediaName, &item.MediaType, &item.MediaURL, &item.Duration, &item.Position, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Media = &models.Media{ID: item.MediaID, Name: mediaName, Type: item.MediaType, URL: item.MediaURL}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) GetPlaylistItemByID(itemID int) (*models.PlaylistItem, error) {
	row := r.db.QueryRow(`
		SELECT p.id, p.window_id, p.media_id, m.name, m.type, m.url, p.duration, p.position, p.created_at, p.updated_at
		FROM playlist_items p
		JOIN media m ON m.id = p.media_id
		WHERE p.id = ?`, itemID)
	var item models.PlaylistItem
	var mediaName string
	if err := row.Scan(&item.ID, &item.WindowID, &item.MediaID, &mediaName, &item.MediaType, &item.MediaURL, &item.Duration, &item.Position, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	item.Media = &models.Media{ID: item.MediaID, Name: mediaName, Type: item.MediaType, URL: item.MediaURL}
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

func (r *Repository) CreateSyncEvent(mediaID, duration int, start time.Time) (*models.SyncEvent, error) {
	if _, err := r.db.Exec("UPDATE sync_events SET status = 'expired' WHERE status = 'active'"); err != nil {
		return nil, err
	}
	endTime := start.Add(time.Duration(duration) * time.Second)
	res, err := r.db.Exec("INSERT INTO sync_events(media_id, start_time, duration, end_time, status, created_at) VALUES (?, ?, ?, ?, ?, ?)", mediaID, start.UTC().Format(time.RFC3339), duration, endTime.UTC().Format(time.RFC3339), "active", time.Now().UTC().Format(time.RFC3339))
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
		SELECT s.id, s.media_id, m.name, m.type, m.url, s.start_time, s.duration, s.end_time, s.status, s.created_at
		FROM sync_events s
		JOIN media m ON m.id = s.media_id
		WHERE s.id = ?`, id)
	var event models.SyncEvent
	var startTime, endTime, createdAt string
	var mediaName string
	if err := row.Scan(&event.ID, &event.MediaID, &mediaName, &event.MediaType, &event.MediaURL, &startTime, &event.Duration, &endTime, &event.Status, &createdAt); err != nil {
		return nil, err
	}
	event.Media = &models.Media{ID: event.MediaID, Name: mediaName, Type: event.MediaType, URL: event.MediaURL}
	if parsed, err := time.Parse(time.RFC3339, startTime); err == nil {
		event.StartTime = parsed
	}
	if parsed, err := time.Parse(time.RFC3339, endTime); err == nil {
		event.EndTime = parsed
	}
	event.CreatedAt = createdAt
	return &event, nil
}

func (r *Repository) GetCurrentSyncEvent() (*models.SyncEvent, error) {
	row := r.db.QueryRow(`
		SELECT s.id, s.media_id, m.name, m.type, m.url, s.start_time, s.duration, s.end_time, s.status, s.created_at
		FROM sync_events s
		JOIN media m ON m.id = s.media_id
		WHERE s.status = 'active'
		ORDER BY s.created_at DESC
		LIMIT 1`)
	var event models.SyncEvent
	var startTime, endTime, createdAt string
	var mediaName string
	if err := row.Scan(&event.ID, &event.MediaID, &mediaName, &event.MediaType, &event.MediaURL, &startTime, &event.Duration, &endTime, &event.Status, &createdAt); err != nil {
		return nil, err
	}
	event.Media = &models.Media{ID: event.MediaID, Name: mediaName, Type: event.MediaType, URL: event.MediaURL}
	if parsed, err := time.Parse(time.RFC3339, startTime); err == nil {
		event.StartTime = parsed
	}
	if parsed, err := time.Parse(time.RFC3339, endTime); err == nil {
		event.EndTime = parsed
	}
	event.CreatedAt = createdAt
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

func (r *Repository) CreateDefaultMedia() error {
	seed := []struct {
		name      string
		mediaType string
		url       string
		duration  int
	}{
		{"M1", "image", "https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=1200&q=80", 5},
		{"M2", "image", "https://images.unsplash.com/photo-1493246507139-91e8fad9978e?auto=format&fit=crop&w=1200&q=80", 6},
		{"M3", "video", "https://interactive-examples.mdn.mozilla.net/media/cc0-videos/flower.mp4", 8},
		{"M4", "image", "https://images.unsplash.com/photo-1470770841072-f978cf4d019e?auto=format&fit=crop&w=1200&q=80", 4},
		{"M5", "image", "https://images.unsplash.com/photo-1501785888041-af3ef285b470?auto=format&fit=crop&w=1200&q=80", 7},
		{"M6", "video", "https://www.w3schools.com/html/mov_bbb.mp4", 9},
		{"Blank", "blank", "", 4},
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
