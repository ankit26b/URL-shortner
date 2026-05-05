package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"url-shortener/internal/model"
)

type URLRepository struct {
	db *pgxpool.Pool
}

func NewURLRepository(db *pgxpool.Pool) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) FindByLongURL(ctx context.Context, longURL string) (*model.URLMapping, error) {
	const query = `SELECT id, long_url, short_code, expires_at FROM urls WHERE long_url = $1`

	var m model.URLMapping
	err := r.db.QueryRow(ctx, query, longURL).Scan(&m.ID, &m.LongURL, &m.ShortCode, &m.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *URLRepository) FindByShortCode(ctx context.Context, shortCode string) (*model.URLMapping, error) {
	const query = `SELECT id, long_url, short_code, expires_at FROM urls WHERE short_code = $1`

	var m model.URLMapping
	err := r.db.QueryRow(ctx, query, shortCode).Scan(&m.ID, &m.LongURL, &m.ShortCode, &m.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *URLRepository) Insert(ctx context.Context, longURL, shortCode string, expiresAt *time.Time) (*model.URLMapping, error) {
	const query = `INSERT INTO urls (long_url, short_code, expires_at) VALUES ($1, $2, $3) RETURNING id, long_url, short_code, expires_at`

	var m model.URLMapping
	err := r.db.QueryRow(ctx, query, longURL, shortCode, expiresAt).Scan(&m.ID, &m.LongURL, &m.ShortCode, &m.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
