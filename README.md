# Multi-Window Media Sequencer with Sync Playback

A full-stack media scheduler for coordinating playlists across multiple display windows with a global sync event system.

## Features

- Multi-window dashboard for managing separate media playlists
- Media library with image, video, and blank asset types
- Per-window playback queue with timing metadata
- Global sync action that triggers the same media across all windows at once
- SQLite-backed persistence and seed data for defaults
- React + Vite frontend and Go REST backend

## Stack

- Backend: Go, SQLite, standard library HTTP server
- Frontend: React, Vite
- Data model: synchronized event-based playback with start/end timestamps

## Local development

### Backend

```bash
cd backend
go mod download
go run ./cmd/server
```

The API listens on http://localhost:8080.

### Frontend

```bash
cd frontend
npm install
npm run dev -- --host 0.0.0.0 --port 5173
```

The UI listens on http://localhost:5173.

### API overview

- GET /api/windows
- POST /api/windows
- GET /api/windows/:id
- DELETE /api/windows/:id
- GET /api/windows/:id/playlist
- POST /api/windows/:id/playlist
- PUT /api/windows/:id/playlist/:itemId
- DELETE /api/windows/:id/playlist/:itemId
- GET /api/media
- POST /api/media
- GET /api/sync
- POST /api/sync
- GET /api/sync/active
- POST /api/sync/stop

## Docker

```bash
docker compose up --build
```

This starts the backend on port 8080 and the frontend on port 5173.

## Default windows and media

The server seeds three display windows and a library of default media assets so the app is ready to use immediately.
