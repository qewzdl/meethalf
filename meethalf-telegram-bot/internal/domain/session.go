package domain

import "time"

type Session struct {
	UserID   int64
	ChatID   int64
	LastSeen time.Time
}
