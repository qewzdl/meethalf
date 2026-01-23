package domain

type MatchAction string

const (
	MatchActionLike    MatchAction = "like"
	MatchActionDislike MatchAction = "dislike"
	MatchActionReport  MatchAction = "report"
)

type MatchCandidate struct {
	Profile     Profile
	HasPrevious bool
}

type MatchActionResult struct {
	Matched bool
}
