package routes

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"multi-screen-media-scheduler/internal/middleware"
	"multi-screen-media-scheduler/internal/services"
)

type API struct {
	service *services.Service
}

func SetupRoutes(service *services.Service) http.Handler {
	api := &API{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/windows", api.handleWindows)
	mux.HandleFunc("/api/windows/", api.handleWindowDetail)
	mux.HandleFunc("/api/media", api.handleMedia)
	mux.HandleFunc("/api/sync", api.handleSync)
	mux.HandleFunc("/api/sync/", api.handleSync)
	return middleware.CorsMiddleware(mux)
}

func (a *API) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func (a *API) writeError(w http.ResponseWriter, status int, message string) {
	a.writeJSON(w, status, map[string]any{"error": message})
}

func (a *API) handleWindows(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		windows, err := a.service.Windows()
		if err != nil {
			a.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.writeJSON(w, http.StatusOK, map[string]any{"windows": windows})
	case http.MethodPost:
		var payload struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			a.writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		window, err := a.service.CreateWindow(payload.Name)
		if err != nil {
			a.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.writeJSON(w, http.StatusCreated, window)
	default:
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *API) handleWindowDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/windows/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		a.writeError(w, http.StatusNotFound, "window not found")
		return
	}

	windowID, err := strconv.Atoi(parts[0])
	if err != nil {
		a.writeError(w, http.StatusBadRequest, "invalid window id")
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			window, err := a.service.WindowByID(windowID)
			if err != nil {
				a.writeError(w, http.StatusNotFound, "window not found")
				return
			}
			a.writeJSON(w, http.StatusOK, window)
		case http.MethodDelete:
			if err := a.service.DeleteWindow(windowID); err != nil {
				a.writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			a.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
		default:
			a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(parts) == 2 && parts[1] == "playlist" {
		switch r.Method {
		case http.MethodGet:
			items, err := a.service.PlaylistForWindow(windowID)
			if err != nil {
				a.writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			a.writeJSON(w, http.StatusOK, map[string]any{"window_id": windowID, "playlist": items})
		case http.MethodPost:
			var payload struct {
				MediaID  int `json:"media_id"`
				Duration int `json:"duration"`
				Position int `json:"position"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				a.writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}
			item, err := a.service.AddPlaylistItem(windowID, payload.MediaID, payload.Duration, payload.Position)
			if err != nil {
				a.writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			a.writeJSON(w, http.StatusCreated, item)
		default:
			a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(parts) == 3 && parts[1] == "playlist" {
		itemID, err := strconv.Atoi(parts[2])
		if err != nil {
			a.writeError(w, http.StatusBadRequest, "invalid item id")
			return
		}
		switch r.Method {
		case http.MethodPut:
			var payload struct {
				Duration int `json:"duration"`
				Position int `json:"position"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				a.writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}
			item, err := a.service.UpdatePlaylistItem(itemID, payload.Duration, payload.Position)
			if err != nil {
				a.writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			a.writeJSON(w, http.StatusOK, item)
		case http.MethodDelete:
			if err := a.service.DeletePlaylistItem(itemID); err != nil {
				a.writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			a.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
		default:
			a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	a.writeError(w, http.StatusNotFound, "resource not found")
}

func (a *API) handleMedia(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.service.Media()
		if err != nil {
			a.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.writeJSON(w, http.StatusOK, map[string]any{"media": items})
	case http.MethodPost:
		var payload struct {
			Name            string `json:"name"`
			Type            string `json:"type"`
			URL             string `json:"url"`
			DefaultDuration int    `json:"default_duration"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			a.writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		media, err := a.service.CreateMedia(payload.Name, payload.Type, payload.URL, payload.DefaultDuration)
		if err != nil {
			a.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		a.writeJSON(w, http.StatusCreated, media)
	default:
		a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *API) handleSync(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/sync")
	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			event, err := a.service.ActiveSync()
			if err != nil {
				a.writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			a.writeJSON(w, http.StatusOK, map[string]any{"sync": event})
		case http.MethodPost:
			var payload struct {
				MediaID  int `json:"media_id"`
				Duration int `json:"duration"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				a.writeError(w, http.StatusBadRequest, "invalid request body")
				return
			}
			event, err := a.service.CreateSync(payload.MediaID, payload.Duration)
			if err != nil {
				a.writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			a.writeJSON(w, http.StatusCreated, map[string]any{"sync": event})
		default:
			a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}
	if path == "/active" {
		event, err := a.service.ActiveSync()
		if err != nil {
			a.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if event == nil {
			a.writeJSON(w, http.StatusOK, map[string]any{"sync": nil})
			return
		}
		event.Status = "active"
		if time.Now().UTC().After(event.EndTime) {
			event.Status = "expired"
		}
		a.writeJSON(w, http.StatusOK, map[string]any{"sync": event})
		return
	}
	if path == "/stop" {
		if r.Method != http.MethodPost {
			a.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if err := a.service.StopSync(); err != nil {
			a.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		a.writeJSON(w, http.StatusOK, map[string]any{"status": "stopped"})
		return
	}
	a.writeError(w, http.StatusNotFound, "sync route not found")
}
