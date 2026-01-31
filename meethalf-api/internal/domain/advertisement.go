package domain

import "time"

type Advertisement struct {
	ID        int64
	Text      string
	PhotoID   string
	CreatedAt time.Time
}
