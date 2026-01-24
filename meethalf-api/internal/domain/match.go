package domain

import "time"

type MatchAction string

const (
	MatchActionLike    MatchAction = "like"
	MatchActionDislike MatchAction = "dislike"
	MatchActionReport  MatchAction = "report"
)

type MatchSession struct {
	ViewerID       int64
	GenderFilter   Gender
	Accuracy       int
	SessionVersion int64
	CurrentIndex   int
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
