package domain

import "time"

const (
	CommandStart                = "start"
	CommandProfile              = "profile"
	CommandProfileView          = "profile_view"
	CommandProfilePreview       = "profile_preview"
	CommandProfileEdit          = "profile_edit"
	CommandProfileEditName      = "profile_edit_name"
	CommandProfileEditGender    = "profile_edit_gender"
	CommandProfileEditBirthDate = "profile_edit_birth_date"
	CommandProfileEditCountry   = "profile_edit_country"
	CommandProfileEditCity      = "profile_edit_city"
	CommandProfileEditDesc      = "profile_edit_desc"
	CommandProfileEditEmoji     = "profile_edit_emoji"
	CommandProfileEditPhotos    = "profile_edit_photos"
	CommandProfileSettings      = "profile_settings"
	CommandProfileDelete        = "profile_delete"
	CommandProfileDeleteConfirm = "profile_delete_confirm"
	CommandProfileDeleteCancel  = "profile_delete_cancel"
)

type IncomingMessage struct {
	ChatID     int64
	User       User
	Text       string
	Command    string
	Arguments  string
	PhotoIDs   []string
	ReceivedAt time.Time
}

type InlineButton struct {
	Text         string
	CallbackData string
}

type InlineKeyboard struct {
	Buttons [][]InlineButton
}

type OutgoingMessage struct {
	ChatID          int64
	Text            string
	ParseMode       string
	DisablePreview  bool
	InlineKeyboard  *InlineKeyboard
	CallbackQueryID string
	CallbackText    string
	PhotoIDs        []string
}
