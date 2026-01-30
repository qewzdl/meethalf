package matching

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"meethalf-api/internal/domain"
)

const (
	minAccuracy         = 0
	maxAccuracy         = 4
	defaultAgeWindow    = 3
	defaultHistoryLimit = 20
	maxHistoryLimit     = 100
	teenMinAge          = 16
	teenMaxAge          = 17
	adultMinAge         = 18
)

var (
	ErrInvalidUserID    = errors.New("user id is required")
	ErrInvalidGender    = errors.New("gender filter is invalid")
	ErrInvalidAccuracy  = errors.New("accuracy must be between 0 and 4")
	ErrInvalidAction    = errors.New("match action is invalid")
	ErrInvalidQuery     = errors.New("search query is required")
	ErrProfileNotFound  = errors.New("profile not found")
	ErrSessionNotFound  = errors.New("search session not found")
	ErrNoCandidates     = errors.New("no candidates found")
	ErrNoPrevious       = errors.New("no previous candidate")
	ErrInvalidTargetID  = errors.New("target user id is required")
	ErrInvalidSelfMatch = errors.New("viewer and target must be different")
	ErrUserBanned       = errors.New("user is banned")
	ErrAIUnavailable    = errors.New("ai search is not configured")
)

type Usecase interface {
	Start(ctx context.Context, viewerID int64, gender domain.Gender, accuracy int) (domain.MatchCandidate, error)
	SearchAI(ctx context.Context, viewerID int64, message string) (domain.MatchCandidate, error)
	Next(ctx context.Context, viewerID int64) (domain.MatchCandidate, error)
	Previous(ctx context.Context, viewerID int64) (domain.MatchCandidate, error)
	History(ctx context.Context, viewerID int64, limit, offset int) (HistoryList, error)
	ResetChoices(ctx context.Context, viewerID int64) error
	RecordAction(ctx context.Context, viewerID, targetID int64, action domain.MatchAction) (domain.MatchActionResult, error)
	PendingLikes(ctx context.Context, userID int64) ([]domain.Profile, error)
}

type Repository interface {
	GetProfile(ctx context.Context, userID int64) (domain.Profile, error)
	GetSession(ctx context.Context, viewerID int64) (domain.MatchSession, bool, error)
	SaveSession(ctx context.Context, session domain.MatchSession) error
	GetHistoryCandidate(ctx context.Context, viewerID int64, sessionVersion int64, position int) (domain.Profile, bool, error)
	SaveHistoryCandidate(ctx context.Context, viewerID int64, sessionVersion int64, position int, candidateID int64) error
	ListHistory(ctx context.Context, viewerID int64, limit, offset int) ([]domain.MatchHistoryItem, int, error)
	ResetChoices(ctx context.Context, viewerID int64) error
	FindCandidate(ctx context.Context, params CandidateParams) (domain.Profile, bool, error)
	FindAICandidate(ctx context.Context, params AICandidateParams) (domain.Profile, bool, error)
	RecordAction(ctx context.Context, viewerID, targetID int64, action domain.MatchAction) error
	HasAction(ctx context.Context, viewerID, targetID int64, action domain.MatchAction) (bool, error)
	ListPendingLikes(ctx context.Context, userID int64) ([]domain.Profile, []int64, error)
	MarkLikesNotified(ctx context.Context, userID int64, likerIDs []int64) error
}

type CandidateParams struct {
	ViewerID       int64
	GenderFilter   domain.Gender
	Accuracy       int
	SessionVersion int64
	ViewerCountry  domain.Country
	ViewerCity     string
	ViewerAge      int
	ViewerEmoji    domain.ProfileEmojiCode
	AgeWindow      int
	MinAge         *int
	MaxAge         *int
}

type HistoryList struct {
	Items  []domain.MatchHistoryItem
	Total  int
	Limit  int
	Offset int
}

type service struct {
	repo    Repository
	planner searchPlanner
	ai      AIAnalyzer
}

func New(repo Repository, ai AIAnalyzer) Usecase {
	return &service{
		repo:    repo,
		planner: newSearchPlanner(),
		ai:      ai,
	}
}

func (s *service) Start(ctx context.Context, viewerID int64, gender domain.Gender, accuracy int) (domain.MatchCandidate, error) {
	if s == nil || s.repo == nil {
		return domain.MatchCandidate{}, errors.New("matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.MatchCandidate{}, err
	}
	if viewerID <= 0 {
		return domain.MatchCandidate{}, ErrInvalidUserID
	}

	normalizedGender, err := normalizeGenderFilter(gender)
	if err != nil {
		return domain.MatchCandidate{}, err
	}
	if accuracy < minAccuracy || accuracy > maxAccuracy {
		return domain.MatchCandidate{}, ErrInvalidAccuracy
	}

	viewer, err := s.viewerProfile(ctx, viewerID)
	if err != nil {
		return domain.MatchCandidate{}, err
	}

	sessionVersion := int64(1)
	if existing, found, err := s.repo.GetSession(ctx, viewerID); err != nil {
		return domain.MatchCandidate{}, err
	} else if found {
		sessionVersion = existing.SessionVersion + 1
	}

	now := time.Now().UTC()
	session := domain.MatchSession{
		ViewerID:       viewerID,
		GenderFilter:   normalizedGender,
		Accuracy:       accuracy,
		SessionVersion: sessionVersion,
		CurrentIndex:   0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return domain.MatchCandidate{}, err
	}

	return s.nextCandidate(ctx, viewer, session)
}

func (s *service) Next(ctx context.Context, viewerID int64) (domain.MatchCandidate, error) {
	if s == nil || s.repo == nil {
		return domain.MatchCandidate{}, errors.New("matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.MatchCandidate{}, err
	}
	if viewerID <= 0 {
		return domain.MatchCandidate{}, ErrInvalidUserID
	}

	session, found, err := s.repo.GetSession(ctx, viewerID)
	if err != nil {
		return domain.MatchCandidate{}, err
	}
	if !found {
		return domain.MatchCandidate{}, ErrSessionNotFound
	}

	viewer, err := s.viewerProfile(ctx, viewerID)
	if err != nil {
		return domain.MatchCandidate{}, err
	}

	return s.nextCandidate(ctx, viewer, session)
}

func (s *service) Previous(ctx context.Context, viewerID int64) (domain.MatchCandidate, error) {
	if s == nil || s.repo == nil {
		return domain.MatchCandidate{}, errors.New("matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.MatchCandidate{}, err
	}
	if viewerID <= 0 {
		return domain.MatchCandidate{}, ErrInvalidUserID
	}

	session, found, err := s.repo.GetSession(ctx, viewerID)
	if err != nil {
		return domain.MatchCandidate{}, err
	}
	if !found {
		return domain.MatchCandidate{}, ErrSessionNotFound
	}
	if _, err := s.viewerProfile(ctx, viewerID); err != nil {
		return domain.MatchCandidate{}, err
	}
	if session.CurrentIndex <= 1 {
		return domain.MatchCandidate{}, ErrNoPrevious
	}

	position := session.CurrentIndex - 1
	candidate, found, err := s.repo.GetHistoryCandidate(ctx, viewerID, session.SessionVersion, position)
	if err != nil {
		return domain.MatchCandidate{}, err
	}
	if !found {
		return domain.MatchCandidate{}, ErrNoPrevious
	}

	now := time.Now().UTC()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.CurrentIndex = position
	session.UpdatedAt = now
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return domain.MatchCandidate{}, err
	}

	return domain.MatchCandidate{
		Profile:     candidate,
		Position:    position,
		HasPrevious: position > 1,
	}, nil
}

func (s *service) History(ctx context.Context, viewerID int64, limit, offset int) (HistoryList, error) {
	if s == nil || s.repo == nil {
		return HistoryList{}, errors.New("matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return HistoryList{}, err
	}
	if viewerID <= 0 {
		return HistoryList{}, ErrInvalidUserID
	}

	normalizedLimit := limit
	if normalizedLimit <= 0 {
		normalizedLimit = defaultHistoryLimit
	}
	if normalizedLimit > maxHistoryLimit {
		normalizedLimit = maxHistoryLimit
	}

	normalizedOffset := offset
	if normalizedOffset < 0 {
		normalizedOffset = 0
	}

	if _, err := s.viewerProfile(ctx, viewerID); err != nil {
		return HistoryList{}, err
	}

	items, total, err := s.repo.ListHistory(ctx, viewerID, normalizedLimit, normalizedOffset)
	if err != nil {
		return HistoryList{}, err
	}

	return HistoryList{
		Items:  items,
		Total:  total,
		Limit:  normalizedLimit,
		Offset: normalizedOffset,
	}, nil
}

func (s *service) ResetChoices(ctx context.Context, viewerID int64) error {
	if s == nil || s.repo == nil {
		return errors.New("matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if viewerID <= 0 {
		return ErrInvalidUserID
	}

	if err := s.ensureUserExists(ctx, viewerID); err != nil {
		return err
	}

	return s.repo.ResetChoices(ctx, viewerID)
}

func (s *service) RecordAction(ctx context.Context, viewerID, targetID int64, action domain.MatchAction) (domain.MatchActionResult, error) {
	if s == nil || s.repo == nil {
		return domain.MatchActionResult{}, errors.New("matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return domain.MatchActionResult{}, err
	}
	if viewerID <= 0 {
		return domain.MatchActionResult{}, ErrInvalidUserID
	}
	if targetID <= 0 {
		return domain.MatchActionResult{}, ErrInvalidTargetID
	}
	if viewerID == targetID {
		return domain.MatchActionResult{}, ErrInvalidSelfMatch
	}

	if _, err := s.viewerProfile(ctx, viewerID); err != nil {
		return domain.MatchActionResult{}, err
	}

	normalized, err := normalizeAction(action)
	if err != nil {
		return domain.MatchActionResult{}, err
	}

	if err := s.repo.RecordAction(ctx, viewerID, targetID, normalized); err != nil {
		return domain.MatchActionResult{}, err
	}

	if normalized != domain.MatchActionLike {
		return domain.MatchActionResult{}, nil
	}

	matched, err := s.repo.HasAction(ctx, targetID, viewerID, domain.MatchActionLike)
	if err != nil {
		return domain.MatchActionResult{}, err
	}

	result := domain.MatchActionResult{Matched: matched}
	if matched {
		notifyErr := errors.Join(
			s.repo.MarkLikesNotified(ctx, viewerID, []int64{targetID}),
			s.repo.MarkLikesNotified(ctx, targetID, []int64{viewerID}),
		)
		if notifyErr != nil {
			return result, notifyErr
		}
	}

	return result, nil
}

func (s *service) PendingLikes(ctx context.Context, userID int64) ([]domain.Profile, error) {
	if s == nil || s.repo == nil {
		return nil, errors.New("matching repository is not configured")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}

	if _, err := s.viewerProfile(ctx, userID); err != nil {
		return nil, err
	}

	likes, likerIDs, err := s.repo.ListPendingLikes(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(likerIDs) == 0 {
		return likes, nil
	}

	if err := s.repo.MarkLikesNotified(ctx, userID, likerIDs); err != nil {
		return nil, err
	}

	return likes, nil
}

func (s *service) nextCandidate(ctx context.Context, viewer domain.Profile, session domain.MatchSession) (domain.MatchCandidate, error) {
	position := session.CurrentIndex + 1
	if position < 1 {
		position = 1
	}

	candidate, found, err := s.repo.GetHistoryCandidate(ctx, session.ViewerID, session.SessionVersion, position)
	if err != nil {
		return domain.MatchCandidate{}, err
	}
	if !found {
		candidate, found, err = s.findCandidate(ctx, viewer, session)
		if err != nil {
			return domain.MatchCandidate{}, err
		}
		if !found {
			return domain.MatchCandidate{}, ErrNoCandidates
		}

		if err := s.repo.SaveHistoryCandidate(ctx, session.ViewerID, session.SessionVersion, position, candidate.UserID); err != nil {
			return domain.MatchCandidate{}, err
		}
	}

	now := time.Now().UTC()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.CurrentIndex = position
	session.UpdatedAt = now
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return domain.MatchCandidate{}, err
	}

	return domain.MatchCandidate{
		Profile:     candidate,
		Position:    position,
		HasPrevious: position > 1,
	}, nil
}

func (s *service) findCandidate(ctx context.Context, viewer domain.Profile, session domain.MatchSession) (domain.Profile, bool, error) {
	attempts := s.planner.Attempts(session.Accuracy)
	if len(attempts) == 0 {
		return domain.Profile{}, false, nil
	}

	params := CandidateParams{
		ViewerID:       session.ViewerID,
		GenderFilter:   session.GenderFilter,
		SessionVersion: session.SessionVersion,
		ViewerCountry:  viewer.Country,
		ViewerCity:     viewer.City,
		ViewerAge:      viewer.Age,
		ViewerEmoji:    viewer.EmojiCode,
	}
	params.MinAge, params.MaxAge = candidateAgeBounds(viewer.Age)

	for _, attempt := range attempts {
		params.Accuracy = attempt.Accuracy
		params.AgeWindow = attempt.AgeWindow
		candidate, found, err := s.repo.FindCandidate(ctx, params)
		if err != nil {
			return domain.Profile{}, false, err
		}
		if found {
			return candidate, true, nil
		}
	}

	return domain.Profile{}, false, nil
}

func candidateAgeBounds(viewerAge int) (*int, *int) {
	if viewerAge >= adultMinAge {
		minAge := adultMinAge
		return &minAge, nil
	}

	minAge := teenMinAge
	maxAge := teenMaxAge
	return &minAge, &maxAge
}

func normalizeGenderFilter(gender domain.Gender) (domain.Gender, error) {
	normalized := strings.ToLower(strings.TrimSpace(string(gender)))
	if normalized == "" || normalized == "any" || normalized == "all" {
		return domain.GenderUnspecified, nil
	}

	switch domain.Gender(normalized) {
	case domain.GenderMale, domain.GenderFemale, domain.GenderOther, domain.GenderUnspecified:
		return domain.Gender(normalized), nil
	default:
		return "", ErrInvalidGender
	}
}

func normalizeAction(action domain.MatchAction) (domain.MatchAction, error) {
	normalized := strings.ToLower(strings.TrimSpace(string(action)))
	switch domain.MatchAction(normalized) {
	case domain.MatchActionLike, domain.MatchActionDislike, domain.MatchActionReport:
		return domain.MatchAction(normalized), nil
	default:
		return "", ErrInvalidAction
	}
}

func (s *service) viewerProfile(ctx context.Context, userID int64) (domain.Profile, error) {
	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Profile{}, ErrProfileNotFound
		}
		return domain.Profile{}, err
	}

	if profile.IsBanned {
		return domain.Profile{}, ErrUserBanned
	}

	return profile, nil
}

func (s *service) ensureUserExists(ctx context.Context, userID int64) error {
	if s == nil || s.repo == nil {
		return errors.New("matching repository is not configured")
	}

	profile, err := s.repo.GetProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrProfileNotFound
		}
		return err
	}
	if profile.UserID == 0 {
		return ErrProfileNotFound
	}

	return nil
}
