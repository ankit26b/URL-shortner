package repository

import (
	"context"
	"time"

	"url-shortener/internal/model"
)

// URLStore abstracts URL persistence to keep services stateless and shard-ready.
type URLStore interface {
	FindByLongURL(ctx context.Context, longURL string) (*model.URLMapping, error)
	FindByShortCode(ctx context.Context, shortCode string) (*model.URLMapping, error)
	Insert(ctx context.Context, longURL, shortCode string, expiresAt *time.Time) (*model.URLMapping, error)
}

// ClickStore abstracts analytics persistence for future sharded backends.
type ClickStore interface {
	Insert(ctx context.Context, event model.ClickEvent) error
}
