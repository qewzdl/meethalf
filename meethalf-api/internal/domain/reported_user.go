package domain

import "time"

type ReportedUserSummary struct {
	UserID      int64     `json:"user_id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	IsHidden    bool      `json:"is_hidden"`
	IsBanned    bool      `json:"is_banned"`
	IsModerator bool      `json:"is_moderator"`
	ReportCount int       `json:"report_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
