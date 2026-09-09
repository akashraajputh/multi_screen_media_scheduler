# Multi-Window Media Sequencer with Sync Playback

A full-stack media scheduler for coordinating playlists across multiple display windows with a global sync event system.

## Overview

The application runs a Go backend with SQLite persistence and a Vite + React frontend. Each display window maintains its own playlist sequence, and a global synchronization event can temporarily override the playlist across all windows for a defined duration.

## Features

- Multi-window dashboard for managing separate media playlists
- Media library with image, video, and blank asset types
- Per-window playback queue with timing metadata
- Global sync action that triggers the same media across all windows at once
- SQLite-backed persistence and seed data for defaults
- React + Vite frontend and Go REST backend
- Server-timestamp-based sync state so all clients calculate the same overlay window

## Stack

- Backend: Go, SQLite, standard library HTTP server
- Frontend: React, Vite
- Data model: synchronized event-based playback with start/end timestamps

## Architecture

- Backend: Go REST API with layered repository/service design
- Frontend: React dashboard with polling refresh for dynamic data updates
- Persistence: SQLite with schema initialization and seed data

## Project Structure

- backend/
  - cmd/server/
  - internal/database/
  - internal/models/
  - internal/repositories/
  - internal/services/
  - internal/middleware/
  - routes/
  - seed/
- frontend/
  - src/
- docker-compose.yml
- README.md

## Database Schema

The application initializes these tables automatically:

- windows
  - id
  - name
  - created_at
  - updated_at
- media
  - id
  - name
  - type
  - url
  - default_duration
  - created_at
  - updated_at
- playlist_items
  - id
  - window_id
  - media_id
  - duration
  - position
  - created_at
  - updated_at
- sync_events
  - id
  - media_id
  - start_time
  - duration
  - end_time
  - status
  - created_at

## API Documentation

### Windows

- GET /api/windows
- POST /api/windows
- GET /api/windows/:id
- DELETE /api/windows/:id

### Playlist

- GET /api/windows/:id/playlist
- POST /api/windows/:id/playlist
- PUT /api/windows/:id/playlist/:itemId
- DELETE /api/windows/:id/playlist/:itemId

### Media

- GET /api/media
- POST /api/media

### Synchronization

- GET /api/sync
- POST /api/sync
- GET /api/sync/active
- POST /api/sync/stop

## 5-Hour Cycle Logic

Each window loads its configured playlist and treats the cycle as five hours. The application does not inject blank playback merely because the configured items are short; blank playback appears only when blank is explicitly included as a playlist item.

## Synchronization Architecture

The backend creates a sync event with a server timestamp, duration, and end time. Clients compare the current server time against the sync start time and display the synced media while they are still within the active sync window. Once the duration expires, the window resumes its standard playlist.

Server timestamps are used to reduce timing drift between browser clients, which is important when several windows must appear synchronized.

## Dynamic Playlist Updates

Playlist items are persisted to SQLite and fetched per window. The frontend refreshes the list and displays the updated queue without requiring a full app restart.

## Local Setup

### Backend

```bash
cd backend
go mod download
go run ./cmd/server
```

The backend listens on http://localhost:8080.

### Frontend

```bash
cd frontend
npm install
npm run dev -- --host 0.0.0.0 --port 5173
```

The frontend listens on http://localhost:5173.

## Environment Variables

- PORT: backend listen port
- DB_PATH: SQLite database path
- VITE_API_URL: frontend API base URL

## Docker

```bash
docker compose up --build
```

This starts the backend on port 8080 and the frontend on port 5173.

## Seed Data

The server seeds three display windows and a library of default media assets so the app is ready to use immediately.

## Testing

### Backend

```bash
cd backend
go test ./...
```

### Frontend

```bash
cd frontend
npm run build
```

## Assumptions

- The app is intended for local or internally managed deployment.
- The latest sync event replaces the previous active event.
- Refreshing the browser during an active sync will resume from the backend state once the UI reloads.

## Known Limitations

- This is a local-first implementation rather than a large multi-user production cluster.
- The update pattern uses polling rather than WebSockets or SSE.
- Media files are loaded from public URLs and depend on browser access to those URLs.

## Live Demo

Local frontend URL:

- http://localhost:5173

Local backend URL:

- http://localhost:8080

## Deployment

This project is structured for Docker-based deployment and is compatible with Render.

### Render deployment architecture

- Backend: Go REST API container on port 8080
- Frontend: React production build served on port 4173
- Database: SQLite stored in the system temp directory on the free plan so the app still runs without a mounted disk
- Communication: frontend calls the backend through VITE_API_URL, not localhost

### Docker files

- backend/Dockerfile
- frontend/Dockerfile
- render.yaml

### Render setup steps

1. Push this project to a GitHub repository.
2. In Render, click New and choose Blueprint.
3. Connect the GitHub repo.
4. Render will detect render.yaml and create both services automatically.
5. Confirm backend service settings:
   - Environment: Docker
   - Port: 8080
   - Environment variables:
     - PORT=8080
     - DB_PATH=/tmp/media_scheduler.db
6. Confirm frontend service settings:
   - Environment: Docker
   - Port: 4173
   - Environment variable:
     - VITE_API_URL=https://<your-backend-render-url>
7. Deploy both services.
8. Open the frontend URL in the browser and verify the app loads.

### Example frontend env var

```env
VITE_API_URL=https://media-scheduler-backend.onrender.com
```

### Example backend env var

```env
PORT=8080
DB_PATH=/tmp/media_scheduler.db
```

### Render notes

- Use the actual backend Render URL in the frontend; do not use localhost in production.
- On the free Render plan, persistent disks are not supported, so SQLite data is stored in /tmp and will reset when the service restarts or redeploys.
- This setup is suitable for demo or testing deployments, not for long-term production persistence.
- The frontend is served as a production build and should not use the Vite dev server in Render.

### Free plan deployment warning

Render free-tier services do not support persistent disk mounts. Because this project uses SQLite, the database will be ephemeral on free hosting. If you need durable data, upgrade to a paid plan or move the database to an external service.

### Local Docker run

```bash
docker compose up --build
```

This starts the backend on port 8080 and the frontend on port 5173 locally.

### Production deployment checklist

- GitHub repo is connected to Render
- Backend is running on port 8080
- Frontend API URL points to the live backend URL
- Persistent disk is mounted to /data
- Frontend loads successfully in the browser

### Troubleshooting

- If the frontend cannot connect to the API, verify VITE_API_URL is set to the deployed backend URL.
- If the backend fails to start, verify the port and DB_PATH environment variables are correct.
- On the free Render plan, database data is expected to reset after restarts or redeploys because the app uses SQLite in the temp filesystem.
- If you need long-term persistence, upgrade to a paid Render plan or move the database to an external service.
