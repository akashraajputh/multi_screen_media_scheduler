package services

import (
	"path/filepath"
	"testing"

	"multi-screen-media-scheduler/internal/database"
)

func newTestService(t *testing.T) *Service {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "scheduler_test.db")
	db, err := database.Initialize(path)
	if err != nil {
		t.Fatalf("initialize test database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	service := NewService(db)
	if err := service.EnsureSeedData(); err != nil {
		t.Fatalf("seed test data: %v", err)
	}

	return service
}

func TestServiceCreateAndListWindows(t *testing.T) {
	service := newTestService(t)

	window, err := service.CreateWindow("Window 4")
	if err != nil {
		t.Fatalf("create window: %v", err)
	}
	if window.Name != "Window 4" {
		t.Fatalf("expected window name Window 4, got %q", window.Name)
	}

	windows, err := service.Windows()
	if err != nil {
		t.Fatalf("list windows: %v", err)
	}
	if len(windows) < 4 {
		t.Fatalf("expected at least 4 windows, got %d", len(windows))
	}
}

func TestServiceCreatesMediaAndPlaylistItem(t *testing.T) {
	service := newTestService(t)

	media, err := service.CreateMedia("Test Clip", "video", "https://example.com/test.mp4", 12)
	if err != nil {
		t.Fatalf("create media: %v", err)
	}
	if media.Name != "Test Clip" {
		t.Fatalf("expected media name Test Clip, got %q", media.Name)
	}

	windows, err := service.Windows()
	if err != nil {
		t.Fatalf("get windows: %v", err)
	}
	if len(windows) == 0 {
		t.Fatal("expected seeded windows")
	}

	windowID := windows[0].ID
	item, err := service.AddPlaylistItem(windowID, media.ID, 12, 0)
	if err != nil {
		t.Fatalf("add playlist item: %v", err)
	}
	if item.WindowID != windowID {
		t.Fatalf("expected item window %d, got %d", windowID, item.WindowID)
	}

	playlist, err := service.PlaylistForWindow(windowID)
	if err != nil {
		t.Fatalf("fetch playlist: %v", err)
	}
	if len(playlist) == 0 {
		t.Fatal("expected playlist items to be present")
	}
}

func TestServiceSyncLifecycle(t *testing.T) {
	service := newTestService(t)

	media, err := service.CreateMedia("Sync Asset", "image", "https://example.com/sync.png", 15)
	if err != nil {
		t.Fatalf("create sync media: %v", err)
	}

	event, err := service.CreateSync(media.ID, 30)
	if err != nil {
		t.Fatalf("create sync: %v", err)
	}
	if event == nil || event.MediaID != media.ID {
		t.Fatal("expected active sync event to be created")
	}

	active, err := service.ActiveSync()
	if err != nil {
		t.Fatalf("get active sync: %v", err)
	}
	if active == nil || active.ID != event.ID {
		t.Fatal("expected created sync to be active")
	}

	if err := service.StopSync(); err != nil {
		t.Fatalf("stop sync: %v", err)
	}

	active, err = service.ActiveSync()
	if err != nil {
		t.Fatalf("get active sync after stop: %v", err)
	}
	if active != nil {
		t.Fatal("expected no active sync after stop")
	}
}

func TestServicePlaylistCycleAndEmptyStateHandling(t *testing.T) {
	service := newTestService(t)
	windows, err := service.Windows()
	if err != nil {
		t.Fatalf("get windows: %v", err)
	}
	if len(windows) == 0 {
		t.Fatal("no windows for test")
	}

	cycle, err := service.ValidatePlaylistCycle(windows[0].ID)
	if err != nil {
		t.Fatalf("validate cycle: %v", err)
	}
	if cycle <= 0 {
		t.Fatal("expected a positive cycle duration")
	}

	state, err := service.PlaylistState(windows[0].ID)
	if err != nil {
		t.Fatalf("get playlist state: %v", err)
	}
	if !state["looped"].(bool) {
		t.Fatal("expected seeded playlist to be looped")
	}
	if _, ok := state["items"]; !ok {
		t.Fatal("expected playlist state to include items")
	}
}
