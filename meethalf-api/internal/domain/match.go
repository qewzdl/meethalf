package domain

import "time"

type MatchAction string

const (
	MatchActionLike    MatchAction = "like"
	MatchActionDislike MatchAction = "dislike"
	MatchActionReport  MatchAction = "report"
)

type MatchingAlgorithmVersion string

const (
	MatchingAlgorithmV1 MatchingAlgorithmVersion = "matching_v1"
	MatchingAlgorithmV2 MatchingAlgorithmVersion = "matching_v2"
)

type MatchSession struct {
	ViewerID         int64
	GenderFilter     Gender
	Accuracy         int
	AlgorithmVersion MatchingAlgorithmVersion
	SessionVersion   int64
	CurrentIndex     int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type MatchCandidate struct {
	Profile     Profile
	Position    int
	HasPrevious bool
}

type MatchHistoryItem struct {
	Profile  Profile
	Position int
	Action   MatchAction
}

type MatchActionResult struct {
	Matched bool
}

type CandidateParams struct {
	ViewerID       int64
	GenderFilter   Gender
	Accuracy       int
	SessionVersion int64
	ViewerCountry  Country
	ViewerCity     string
	ViewerAge      int
	ViewerEmoji    ProfileEmojiCode
	AgeWindow      int
	MinAge         *int
	MaxAge         *int
}

type AICandidateParams struct {
	ViewerID     int64
	GenderFilter Gender
	MinAge       *int
	MaxAge       *int
	Country      Country
	City         string
	EmojiCode    ProfileEmojiCode
	Keywords     []string
}
