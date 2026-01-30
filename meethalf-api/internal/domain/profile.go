package domain

import "time"

type Gender string

const (
	GenderMale        Gender = "male"
	GenderFemale      Gender = "female"
	GenderOther       Gender = "other"
	GenderUnspecified Gender = "unspecified"
)

type Country string

const (
	CountryRussia     Country = "russia"
	CountryKazakhstan Country = "kazakhstan"
	CountryBelarus    Country = "belarus"
)

type ProfileEmojiCode string

const (
	ProfileEmojiLeader        ProfileEmojiCode = "LDR"
	ProfileEmojiStrategist    ProfileEmojiCode = "STR"
	ProfileEmojiAnalyst       ProfileEmojiCode = "ANA"
	ProfileEmojiCreator       ProfileEmojiCode = "CRT"
	ProfileEmojiCommunicator  ProfileEmojiCode = "COM"
	ProfileEmojiEmpath        ProfileEmojiCode = "EMP"
	ProfileEmojiMediator      ProfileEmojiCode = "MED"
	ProfileEmojiPerfectionist ProfileEmojiCode = "PRF"
	ProfileEmojiResearcher    ProfileEmojiCode = "RSR"
	ProfileEmojiInnovator     ProfileEmojiCode = "INN"
	ProfileEmojiExecutor      ProfileEmojiCode = "EXE"
	ProfileEmojiAdventurer    ProfileEmojiCode = "ADV"
	ProfileEmojiContemplator  ProfileEmojiCode = "CNT"
	ProfileEmojiRealist       ProfileEmojiCode = "RLS"
	ProfileEmojiMotivator     ProfileEmojiCode = "MOT"
	ProfileEmojiSkeptic       ProfileEmojiCode = "SKP"
)

type Profile struct {
	UserID         int64            `json:"user_id"`
	Username       string           `json:"username"`
	Name           string           `json:"name"`
	Gender         Gender           `json:"gender"`
	BirthDate      time.Time        `json:"birth_date"`
	Age            int              `json:"age"`
	Country        Country          `json:"country"`
	City           string           `json:"city"`
	Description    string           `json:"description"`
	EmojiCode      ProfileEmojiCode `json:"emoji_code"`
	Photos         []string         `json:"photos"`
	IsHidden       bool             `json:"is_hidden"`
	IsBanned       bool             `json:"is_banned"`
	IsShadowBanned bool             `json:"is_shadow_banned"`
	IsModerator    bool             `json:"is_moderator"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}
