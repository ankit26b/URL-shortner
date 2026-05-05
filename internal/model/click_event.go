package model

import "time"

type ClickEvent struct {
	ShortCode string
	Timestamp time.Time
	IP        string
	UserAgent string
}
