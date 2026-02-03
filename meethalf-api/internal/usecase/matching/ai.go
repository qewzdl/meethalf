package matching

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"meethalf-api/internal/domain"
)

const (
	aiMinAgeAllowed   = 16
	aiMaxAgeAllowed   = 120
	aiMaxKeywords     = 8
	aiMaxKeywordRunes = 32
)

type AIAnalyzer interface {
	Analyze(ctx context.Context, input string) (AIQuery, error)
}

type AIQuery struct {
	Gender    domain.Gender
	MinAge    *int
	MaxAge    *int
	Country   domain.Country
	City      string
	EmojiCode domain.ProfileEmojiCode
	Keywords  []string
}

func (s *service) SearchAI(ctx context.Context, viewerID int64, message string) (domain.MatchCandidate, error) {
	if s == nil || s.repo == nil {
		return domain.MatchCandidate{}, errors.New("matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.MatchCandidate{}, err
	}
	if viewerID <= 0 {
		return domain.MatchCandidate{}, ErrInvalidUserID
	}

	query := strings.TrimSpace(message)
	if query == "" {
		return domain.MatchCandidate{}, ErrInvalidQuery
	}
	if s.ai == nil {
		return domain.MatchCandidate{}, ErrAIUnavailable
	}

	viewer, err := s.viewerProfile(ctx, viewerID)
	if err != nil {
		return domain.MatchCandidate{}, err
	}

	aiQuery, err := s.ai.Analyze(ctx, query)
	if err != nil {
		return domain.MatchCandidate{}, err
	}

	normalized := normalizeAIQuery(aiQuery)
	allowedMin, allowedMax := candidateAgeBounds(viewer.Age)
	minAge, maxAge := mergeAgeBounds(allowedMin, allowedMax, normalized.MinAge, normalized.MaxAge)
	if minAge != nil && maxAge != nil && *minAge > *maxAge {
		return domain.MatchCandidate{}, ErrNoCandidates
	}

	candidate, found, err := s.repo.FindAICandidate(ctx, domain.AICandidateParams{
		ViewerID:     viewerID,
		GenderFilter: normalized.Gender,
		MinAge:       minAge,
		MaxAge:       maxAge,
		Country:      normalized.Country,
		City:         normalized.City,
		EmojiCode:    normalized.EmojiCode,
		Keywords:     normalized.Keywords,
	})
	if err != nil {
		return domain.MatchCandidate{}, err
	}
	if !found {
		return domain.MatchCandidate{}, ErrNoCandidates
	}

	return domain.MatchCandidate{
		Profile:     candidate,
		Position:    1,
		HasPrevious: false,
	}, nil
}

func (s *service) AIAvailable(ctx context.Context, viewerID int64) (bool, error) {
	if s == nil || s.repo == nil {
		return false, errors.New("matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if viewerID <= 0 {
		return false, ErrInvalidUserID
	}
	if _, err := s.viewerProfile(ctx, viewerID); err != nil {
		return false, err
	}
	if s.ai == nil {
		return false, nil
	}
	return true, nil
}

func normalizeAIQuery(query AIQuery) AIQuery {
	gender := query.Gender
	if normalized, err := normalizeGenderFilter(query.Gender); err == nil {
		gender = normalized
	} else {
		gender = domain.GenderUnspecified
	}

	minAge, maxAge := normalizeAgeRange(query.MinAge, query.MaxAge)

	return AIQuery{
		Gender:    gender,
		MinAge:    minAge,
		MaxAge:    maxAge,
		Country:   normalizeCountryFilter(query.Country),
		City:      strings.TrimSpace(query.City),
		EmojiCode: normalizeEmojiFilter(query.EmojiCode),
		Keywords:  normalizeKeywords(query.Keywords),
	}
}

func normalizeAgeRange(minAge, maxAge *int) (*int, *int) {
	normalizedMin := normalizeAgeValue(minAge)
	normalizedMax := normalizeAgeValue(maxAge)
	if normalizedMin != nil && normalizedMax != nil && *normalizedMin > *normalizedMax {
		normalizedMin, normalizedMax = normalizedMax, normalizedMin
	}
	return normalizedMin, normalizedMax
}

func normalizeAgeValue(age *int) *int {
	if age == nil {
		return nil
	}
	value := *age
	if value < aiMinAgeAllowed || value > aiMaxAgeAllowed {
		return nil
	}
	return intPointer(value)
}

func mergeAgeBounds(allowedMin, allowedMax, aiMin, aiMax *int) (*int, *int) {
	min := allowedMin
	if aiMin != nil && (min == nil || *aiMin > *min) {
		min = aiMin
	}

	max := allowedMax
	if aiMax != nil && (max == nil || *aiMax < *max) {
		max = aiMax
	}

	return min, max
}

func normalizeCountryFilter(value domain.Country) domain.Country {
	normalized := strings.ToLower(strings.TrimSpace(string(value)))
	if normalized == "" {
		return ""
	}

	switch domain.Country(normalized) {
	case domain.CountryRussia, domain.CountryKazakhstan, domain.CountryBelarus:
		return domain.Country(normalized)
	default:
		return ""
	}
}

func normalizeEmojiFilter(value domain.ProfileEmojiCode) domain.ProfileEmojiCode {
	normalized := strings.ToUpper(strings.TrimSpace(string(value)))
	if normalized == "" {
		return ""
	}

	switch domain.ProfileEmojiCode(normalized) {
	case domain.ProfileEmojiLeader,
		domain.ProfileEmojiStrategist,
		domain.ProfileEmojiAnalyst,
		domain.ProfileEmojiCreator,
		domain.ProfileEmojiCommunicator,
		domain.ProfileEmojiEmpath,
		domain.ProfileEmojiMediator,
		domain.ProfileEmojiPerfectionist,
		domain.ProfileEmojiResearcher,
		domain.ProfileEmojiInnovator,
		domain.ProfileEmojiExecutor,
		domain.ProfileEmojiAdventurer,
		domain.ProfileEmojiContemplator,
		domain.ProfileEmojiRealist,
		domain.ProfileEmojiMotivator,
		domain.ProfileEmojiSkeptic:
		return domain.ProfileEmojiCode(normalized)
	default:
		return ""
	}
}

func normalizeKeywords(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	unique := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		keyword := strings.ToLower(strings.TrimSpace(value))
		if keyword == "" {
			continue
		}
		if utf8.RuneCountInString(keyword) > aiMaxKeywordRunes {
			continue
		}
		if _, ok := unique[keyword]; ok {
			continue
		}
		unique[keyword] = struct{}{}
		out = append(out, keyword)
		if len(out) >= aiMaxKeywords {
			break
		}
	}

	return out
}

func intPointer(value int) *int {
	return &value
}
