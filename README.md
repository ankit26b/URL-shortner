# URL Shortener Backend (Go + Gin)

Production-ready backend scaffold for a URL shortener service using:
- Go + Gin
- PostgreSQL
- Redis

## Project Structure

```
.
├── cmd/api                # application entrypoint
├── internal/cache/redis   # Redis client initialization
├── internal/config        # env-based configuration
├── internal/handler       # HTTP handlers
├── internal/logger        # logger setup
├── internal/router        # route registration
├── internal/storage/postgres # PostgreSQL connection pool
├── migrations             # database migrations
├── Dockerfile
└── docker-compose.yml
```

## Setup

1. Copy `.env.example` to `.env` and adjust values.
2. Start dependencies:
   ```bash
   docker compose up -d postgres redis
   ```
3. Run app:
   ```bash
   go run ./cmd/api
   ```

## Endpoint

- `GET /api/v1/health` → `200 OK`
