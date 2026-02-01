package domain

import "time"

type UserSummary struct {
	UserID         int64
	Username       string
	Name           string
	IsHidden       bool
	IsBanned       bool
	IsShadowBanned bool
	IsModerator    bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UserList struct {
	Users  []UserSummary
	Total  int
	Limit  int
	Offset int
}

type ReportedUserSummary struct {
	UserID         int64
	Username       string
	Name           string
	IsHidden       bool
	IsBanned       bool
	IsShadowBanned bool
	IsModerator    bool
	ReportCount    int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ReportedUserList struct {
	Users  []ReportedUserSummary
	Total  int
	Limit  int
	Offset int
}

type AdminActionType string

const (
	AdminActionBan          AdminActionType = "ban"
	AdminActionUnban        AdminActionType = "unban"
	AdminActionShadowBan    AdminActionType = "shadow_ban"
	AdminActionUnshadowBan  AdminActionType = "shadow_unban"
	AdminActionHideProfile  AdminActionType = "hide_profile"
	AdminActionShowProfile  AdminActionType = "show_profile"
	AdminActionModerator    AdminActionType = "moderator"
	AdminActionUnmoderator  AdminActionType = "unmoderator"
	AdminActionResetChoices AdminActionType = "reset_choices"
	AdminActionResetStart   AdminActionType = "reset_start"
	AdminActionClearReports AdminActionType = "clear_reports"
	AdminActionPostAd       AdminActionType = "post_ad"
	AdminActionPostAdButton AdminActionType = "post_ad_button"
)

type AdminActionState struct {
	UserID      int64
	ChatID      int64
	Action      AdminActionType
	RequestedAt time.Time
	AdText      string
	AdPhotoID   string
	AdButtons   []AdButton
}
