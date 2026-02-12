# GoDockerUI

Go-based Docker dashboard that simulates a Portainer-like UI.

## Features

- Container listing
- Image listing
- Container start/stop/restart actions
- Dashboard counters
- Automatic simulation fallback when Docker daemon is unavailable

## Run locally

```bash
go mod tidy
go run ./cmd/server
```

Open `http://localhost:8080`.

## API

- `GET /api/v1/health`
- `GET /api/v1/dashboard`
- `GET /api/v1/containers`
- `POST /api/v1/containers/:id/start`
- `POST /api/v1/containers/:id/stop`
- `POST /api/v1/containers/:id/restart`
- `GET /api/v1/images`
