package model

import "time"

type URLMapping struct {
	ID        int64
	LongURL   string
	ShortCode string
	ExpiresAt *time.Time
}
