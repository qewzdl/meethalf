package admin

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"meethalf-api/internal/domain"
)

const (
	defaultListLimit = 50
	maxListLimit     = 200
)

type Usecase interface {
	ListUsers(ctx context.Context, limit, offset int, onlyBanned, onlyModerators bool) (UserList, error)
	ListReportedUsers(ctx context.Context, limit, offset int) (ReportedUserList, error)
	GetUser(ctx context.Context, userID int64) (domain.UserSummary, error)
	GetUserByUsername(ctx context.Context, username string) (domain.UserSummary, error)
	BanUser(ctx context.Context, userID int64) error
	BanUserByUsername(ctx context.Context, username string) error
	UnbanUser(ctx context.Context, userID int64) error
	UnbanUserByUsername(ctx context.Context, username string) error
	MakeModerator(ctx context.Context, userID int64) error
	MakeModeratorByUsername(ctx context.Context, username string) error
	RemoveModerator(ctx context.Context, userID int64) error
	RemoveModeratorByUsername(ctx context.Context, username string) error
}

type Repository interface {
	ListUsers(ctx context.Context, limit, offset int, onlyBanned, onlyModerators bool) ([]domain.UserSummary, int, error)
	ListReportedUsers(ctx context.Context, limit, offset int) ([]domain.ReportedUserSummary, int, error)
	GetUserSummary(ctx context.Context, userID int64) (domain.UserSummary, error)
	UpdateBanStatus(ctx context.Context, userID int64, isBanned bool) error
	UpdateModeratorStatus(ctx context.Context, userID int64, isModerator bool) error
	GetUserIDByUsername(ctx context.Context, username string) (int64, error)
}

type UserList struct {
	Users  []domain.UserSummary
	Total  int
	Limit  int
	Offset int
}

type ReportedUserList struct {
	Users  []domain.ReportedUserSummary
	Total  int
	Limit  int
	Offset int
}

type service struct {
	repo Repository
}

func New(repo Repository) Usecase {
	return &service{repo: repo}
}

func (s *service) ListUsers(ctx context.Context, limit, offset int, onlyBanned, onlyModerators bool) (UserList, error) {
	if s == nil || s.repo == nil {
		return UserList{}, errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return UserList{}, err
	}

	normalizedLimit := limit
	if normalizedLimit <= 0 {
		normalizedLimit = defaultListLimit
	}
	if normalizedLimit > maxListLimit {
		normalizedLimit = maxListLimit
	}

	normalizedOffset := offset
	if normalizedOffset < 0 {
		normalizedOffset = 0
	}

	users, total, err := s.repo.ListUsers(ctx, normalizedLimit, normalizedOffset, onlyBanned, onlyModerators)
	if err != nil {
		return UserList{}, err
	}

	return UserList{
		Users:  users,
		Total:  total,
		Limit:  normalizedLimit,
		Offset: normalizedOffset,
	}, nil
}

func (s *service) ListReportedUsers(ctx context.Context, limit, offset int) (ReportedUserList, error) {
	if s == nil || s.repo == nil {
		return ReportedUserList{}, errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return ReportedUserList{}, err
	}

	normalizedLimit := limit
	if normalizedLimit <= 0 {
		normalizedLimit = defaultListLimit
	}
	if normalizedLimit > maxListLimit {
		normalizedLimit = maxListLimit
	}

	normalizedOffset := offset
	if normalizedOffset < 0 {
		normalizedOffset = 0
	}

	users, total, err := s.repo.ListReportedUsers(ctx, normalizedLimit, normalizedOffset)
	if err != nil {
		return ReportedUserList{}, err
	}

	return ReportedUserList{
		Users:  users,
		Total:  total,
		Limit:  normalizedLimit,
		Offset: normalizedOffset,
	}, nil
}

func (s *service) GetUser(ctx context.Context, userID int64) (domain.UserSummary, error) {
	if s == nil || s.repo == nil {
		return domain.UserSummary{}, errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return domain.UserSummary{}, err
	}

	if userID <= 0 {
		return domain.UserSummary{}, ErrInvalidUserID
	}

	user, err := s.repo.GetUserSummary(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.UserSummary{}, ErrUserNotFound
		}
		return domain.UserSummary{}, err
	}

	return user, nil
}

func (s *service) GetUserByUsername(ctx context.Context, username string) (domain.UserSummary, error) {
	if s == nil || s.repo == nil {
		return domain.UserSummary{}, errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return domain.UserSummary{}, err
	}

	userID, err := s.resolveUserIDByUsername(ctx, username)
	if err != nil {
		return domain.UserSummary{}, err
	}

	return s.GetUser(ctx, userID)
}

func (s *service) BanUser(ctx context.Context, userID int64) error {
	if s == nil || s.repo == nil {
		return errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if userID <= 0 {
		return ErrInvalidUserID
	}

	if err := s.repo.UpdateBanStatus(ctx, userID, true); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (s *service) BanUserByUsername(ctx context.Context, username string) error {
	if s == nil || s.repo == nil {
		return errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	userID, err := s.resolveUserIDByUsername(ctx, username)
	if err != nil {
		return err
	}

	return s.BanUser(ctx, userID)
}

func (s *service) UnbanUser(ctx context.Context, userID int64) error {
	if s == nil || s.repo == nil {
		return errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if userID <= 0 {
		return ErrInvalidUserID
	}

	if err := s.repo.UpdateBanStatus(ctx, userID, false); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (s *service) UnbanUserByUsername(ctx context.Context, username string) error {
	if s == nil || s.repo == nil {
		return errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	userID, err := s.resolveUserIDByUsername(ctx, username)
	if err != nil {
		return err
	}

	return s.UnbanUser(ctx, userID)
}

func (s *service) MakeModerator(ctx context.Context, userID int64) error {
	if s == nil || s.repo == nil {
		return errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if userID <= 0 {
		return ErrInvalidUserID
	}

	if err := s.repo.UpdateModeratorStatus(ctx, userID, true); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (s *service) MakeModeratorByUsername(ctx context.Context, username string) error {
	if s == nil || s.repo == nil {
		return errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	userID, err := s.resolveUserIDByUsername(ctx, username)
	if err != nil {
		return err
	}

	return s.MakeModerator(ctx, userID)
}

func (s *service) RemoveModerator(ctx context.Context, userID int64) error {
	if s == nil || s.repo == nil {
		return errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if userID <= 0 {
		return ErrInvalidUserID
	}

	if err := s.repo.UpdateModeratorStatus(ctx, userID, false); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (s *service) RemoveModeratorByUsername(ctx context.Context, username string) error {
	if s == nil || s.repo == nil {
		return errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	userID, err := s.resolveUserIDByUsername(ctx, username)
	if err != nil {
		return err
	}

	return s.RemoveModerator(ctx, userID)
}

func (s *service) resolveUserIDByUsername(ctx context.Context, username string) (int64, error) {
	if s == nil || s.repo == nil {
		return 0, errors.New("admin repository is not configured")
	}

	if err := ctx.Err(); err != nil {
		return 0, err
	}

	normalized, err := normalizeUsername(username)
	if err != nil {
		return 0, err
	}

	userID, err := s.repo.GetUserIDByUsername(ctx, normalized)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}

	if userID <= 0 {
		return 0, ErrUserNotFound
	}

	return userID, nil
}

var (
	ErrInvalidUserID   = errors.New("user id is required")
	ErrInvalidUsername = errors.New("username is required")
	ErrUserNotFound    = errors.New("user not found")
)

func normalizeUsername(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", ErrInvalidUsername
	}
	if strings.HasPrefix(normalized, "@") {
		normalized = strings.TrimPrefix(normalized, "@")
	}
	if normalized == "" {
		return "", ErrInvalidUsername
	}

	return normalized, nil
}
