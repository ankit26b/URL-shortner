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
3. Run migrations in `migrations/`.
4. Run app:
   ```bash
   go run ./cmd/api
   ```

## Endpoints

- `GET /api/v1/health` → `200 OK`
- `POST /api/v1/shorten` → `200 OK`
  - Request body: `{ "long_url": "https://example.com/very/long/path", "expiry_time": "2026-12-31T23:59:59Z" }`
    - `expiry_time` is optional and must be an RFC3339 timestamp in the future.
  - Response body: `{ "short_code": "abc123" }`
- `GET /:short_code` → `302 Found`
  - Redirects to original URL when found and active.
  - Returns `404` for unknown short codes.
  - Returns `410` for expired links.
