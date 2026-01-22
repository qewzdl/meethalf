package domain

import "time"

type ProfileDeletionConfirmation struct {
	UserID      int64
	ChatID      int64
	RequestedAt time.Time
}
