package models

import "time"

type Window struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Media struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	URL             string `json:"url"`
	DefaultDuration int    `json:"default_duration"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type PlaylistItem struct {
	ID        int    `json:"id"`
	WindowID  int    `json:"window_id"`
	MediaID   int    `json:"media_id"`
	MediaType string `json:"media_type,omitempty"`
	MediaURL  string `json:"media_url,omitempty"`
	Duration  int    `json:"duration"`
	Position  int    `json:"position"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Media     *Media `json:"media,omitempty"`
}

type SyncEvent struct {
	ID        int       `json:"id"`
	MediaID   int       `json:"media_id"`
	MediaType string    `json:"media_type,omitempty"`
	MediaURL  string    `json:"media_url,omitempty"`
	StartTime time.Time `json:"start_time"`
	Duration  int       `json:"duration"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"created_at"`
	Media     *Media    `json:"media,omitempty"`
}
