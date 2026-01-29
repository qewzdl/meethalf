package domain

import "time"

type AgeConfirmation struct {
	UserID      int64
	ChatID      int64
	Username    string
	ConfirmedAt time.Time
}
