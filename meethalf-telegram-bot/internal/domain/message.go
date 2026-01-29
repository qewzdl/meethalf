package domain

import "time"

const (
	CommandStart                = "start"
	CommandAgeConfirmYes        = "age_confirm_yes"
	CommandAgeConfirmNo         = "age_confirm_no"
	CommandCancel               = "cancel"
	CommandProfile              = "profile"
	CommandProfileSetupBack     = "profile_setup_back"
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
	CommandProfileLanguage      = "profile_language"
	CommandLanguage             = "language"
	CommandProfileVisibility    = "profile_visibility"
	CommandProfileDelete        = "profile_delete"
	CommandProfileDeleteConfirm = "profile_delete_confirm"
	CommandProfileDeleteCancel  = "profile_delete_cancel"
	CommandSearchStart          = "search_start"
	CommandSearchRefresh        = "search_refresh"
	CommandSearchGender         = "search_gender"
	CommandSearchAccuracy       = "search_accuracy"
	CommandMatchLike            = "match_like"
	CommandMatchDislike         = "match_dislike"
	CommandMatchReport          = "match_report"
	CommandMatchPrevious        = "match_previous"
	CommandMatchViewProfile     = "match_view_profile"
	CommandMatchViewLike        = "match_view_like"
	CommandMatchViewDislike     = "match_view_dislike"
	CommandMatchViewReport      = "match_view_report"
	CommandMatchHistory         = "match_history"
	CommandMatchHistoryView     = "match_history_view"
	CommandMatchHistoryLike     = "match_history_like"
	CommandMatchHistoryDislike  = "match_history_dislike"
	CommandMatchHistoryReport   = "match_history_report"
	CommandAdminMenu            = "admin_menu"
	CommandAdminUsers           = "admin_users"
	CommandAdminBannedUsers     = "admin_banned_users"
	CommandAdminModerators      = "admin_moderators"
	CommandAdminReports         = "admin_reports"
	CommandAdminBan             = "ban"
	CommandAdminUnban           = "unban"
	CommandAdminModerator       = "moderator"
	CommandAdminUnmoderator     = "unmoderator"
	CommandAdminResetChoices    = "reset_choices"
	CommandAdminResetStart      = "reset_start"
	CommandAdminClearReports    = "clear_reports"
)

type OutgoingMessageKind string

const (
	OutgoingMessageKindLikeNotification  OutgoingMessageKind = "like_notification"
	OutgoingMessageKindMatchNotification OutgoingMessageKind = "match_notification"
)

type IncomingMessage struct {
	ChatID     int64
	MessageID  int
	User       User
	Language   Language
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
	ChatID               int64
	Kind                 OutgoingMessageKind
	Text                 string
	ParseMode            string
	DisablePreview       bool
	InlineKeyboard       *InlineKeyboard
	CallbackQueryID      string
	CallbackText         string
	PhotoIDs             []string
	CleanupFromMessageID int
}
