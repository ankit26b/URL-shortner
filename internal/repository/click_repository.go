package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"url-shortener/internal/model"
)

type ClickRepository struct {
	db *pgxpool.Pool
}

func NewClickRepository(db *pgxpool.Pool) *ClickRepository {
	return &ClickRepository{db: db}
}

func (r *ClickRepository) Insert(ctx context.Context, event model.ClickEvent) error {
	const query = `
		INSERT INTO url_clicks (short_code, clicked_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query, event.ShortCode, event.Timestamp, event.IP, event.UserAgent)
	return err
}
