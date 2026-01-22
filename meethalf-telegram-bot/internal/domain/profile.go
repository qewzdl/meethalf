package domain

import "time"

type ProfileDraftStep string

const (
	ProfileDraftStepName        ProfileDraftStep = "name"
	ProfileDraftStepGender      ProfileDraftStep = "gender"
	ProfileDraftStepBirthDate   ProfileDraftStep = "birth_date"
	ProfileDraftStepCountry     ProfileDraftStep = "country"
	ProfileDraftStepCity        ProfileDraftStep = "city"
	ProfileDraftStepDescription ProfileDraftStep = "description"
	ProfileDraftStepEmoji       ProfileDraftStep = "emoji"
	ProfileDraftStepPhotos      ProfileDraftStep = "photos"
)

type ProfileDraftMode string

const (
	ProfileDraftModeCreate ProfileDraftMode = "create"
	ProfileDraftModeEdit   ProfileDraftMode = "edit"
)

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

type ProfileDraft struct {
	UserID      int64
	ChatID      int64
	Step        ProfileDraftStep
	Name        string
	Gender      Gender
	BirthDate   time.Time
	Country     Country
	City        string
	Description string
	EmojiCode   ProfileEmojiCode
	Photos      []string
	Mode        ProfileDraftMode
	UpdatedAt   time.Time
}

type Profile struct {
	UserID      int64
	Name        string
	Gender      Gender
	BirthDate   time.Time
	Age         int
	Country     Country
	City        string
	Description string
	EmojiCode   ProfileEmojiCode
	Photos      []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
