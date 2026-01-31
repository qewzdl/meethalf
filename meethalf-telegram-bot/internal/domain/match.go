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

type MatchHistoryItem struct {
	Profile  Profile
	Position int
	Action   MatchAction
}

type MatchHistoryList struct {
	Items  []MatchHistoryItem
	Total  int
	Limit  int
	Offset int
}

type MatchLikesList struct {
	Items  []Profile
	Total  int
	Limit  int
	Offset int
}

type MatchActionResult struct {
	Matched bool
}
