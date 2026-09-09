package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"multi-screen-media-scheduler/internal/models"
	"multi-screen-media-scheduler/internal/repositories"
)

const fiveHoursInSeconds = 5 * 60 * 60

type Service struct {
	repo *repositories.Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: repositories.NewRepository(db)}
}

func (s *Service) Windows() ([]models.Window, error) {
	return s.repo.GetWindows()
}

func (s *Service) WindowByID(id int) (*models.Window, error) {
	return s.repo.GetWindow(id)
}

func (s *Service) CreateWindow(name string) (*models.Window, error) {
	if name == "" {
		return nil, errors.New("window name is required")
	}
	return s.repo.CreateWindow(name)
}

func (s *Service) DeleteWindow(id int) error {
	return s.repo.DeleteWindow(id)
}

func (s *Service) Media() ([]models.Media, error) {
	return s.repo.GetMedia()
}

func (s *Service) CreateMedia(name, mediaType, url string, defaultDuration int) (*models.Media, error) {
	if name == "" {
		return nil, errors.New("media name is required")
	}
	if mediaType == "" {
		return nil, errors.New("media type is required")
	}
	if defaultDuration <= 0 {
		defaultDuration = 5
	}
	return s.repo.CreateMedia(name, mediaType, url, defaultDuration)
}

func (s *Service) PlaylistForWindow(windowID int) ([]models.PlaylistItem, error) {
	if _, err := s.repo.GetWindow(windowID); err != nil {
		return nil, err
	}
	return s.repo.WindowPlaylist(windowID)
}

func (s *Service) AddPlaylistItem(windowID, mediaID, duration, position int) (*models.PlaylistItem, error) {
	if _, err := s.repo.GetWindow(windowID); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetMediaByID(mediaID); err != nil {
		return nil, err
	}
	if duration <= 0 {
		duration = 5
	}
	return s.repo.AddPlaylistItem(windowID, mediaID, duration, position)
}

func (s *Service) UpdatePlaylistItem(itemID, duration, position int) (*models.PlaylistItem, error) {
	if duration <= 0 {
		duration = 5
	}
	return s.repo.UpdatePlaylistItem(itemID, duration, position)
}

func (s *Service) DeletePlaylistItem(itemID int) error {
	return s.repo.DeletePlaylistItem(itemID)
}

func (s *Service) CreateSync(mediaID, duration int) (*models.SyncEvent, error) {
	if duration <= 0 {
		duration = 30
	}
	if _, err := s.repo.GetMediaByID(mediaID); err != nil {
		return nil, err
	}
	return s.repo.CreateSyncEvent(mediaID, duration, time.Now().UTC())
}

func (s *Service) ActiveSync() (*models.SyncEvent, error) {
	if err := s.repo.ExpireSyncEvents(); err != nil {
		return nil, err
	}
	event, err := s.repo.GetCurrentSyncEvent()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return event, nil
}

func (s *Service) StopSync() error {
	return s.repo.StopSyncEvent()
}

func (s *Service) ValidatePlaylistCycle(windowID int) (int, error) {
	seconds, err := s.repo.WindowCycleDuration(windowID)
	if err != nil {
		return 0, err
	}
	if seconds <= 0 {
		return 0, nil
	}
	if seconds < fiveHoursInSeconds {
		return seconds, nil
	}
	return fiveHoursInSeconds, nil
}

func (s *Service) PlaylistState(windowID int) (map[string]any, error) {
	items, err := s.PlaylistForWindow(windowID)
	if err != nil {
		return nil, err
	}
	totalSeconds := 0
	for _, item := range items {
		totalSeconds += item.Duration
	}
	return map[string]any{
		"window_id":      windowID,
		"total_duration": totalSeconds,
		"cycle_seconds":  fiveHoursInSeconds,
		"items":          items,
		"looped":         totalSeconds > 0,
		"has_blank_only": totalSeconds == 0,
	}, nil
}

func (s *Service) SeedDefaults() error {
	return s.repo.CreateDefaultMedia()
}

func (s *Service) CreateSeedWindows() error {
	windows := []string{"Window 1", "Window 2", "Window 3"}
	for _, name := range windows {
		if _, err := s.repo.CreateWindow(name); err != nil {
			return fmt.Errorf("create seed window %s: %w", name, err)
		}
	}
	return nil
}

func (s *Service) EnsureSeedData() error {
	if err := s.repo.CreateDefaultMedia(); err != nil {
		return err
	}
	windows, err := s.repo.GetWindows()
	if err != nil {
		return err
	}
	if len(windows) >= 3 {
		return nil
	}
	return s.CreateSeedWindows()
}
