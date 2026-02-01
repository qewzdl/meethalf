package domain

import "time"

type Advertisement struct {
	ID        int64
	Text      string
	PhotoID   string
	Buttons   []AdButton
	CreatedAt time.Time
}

type AdButton struct {
	Text string
	URL  string
}
