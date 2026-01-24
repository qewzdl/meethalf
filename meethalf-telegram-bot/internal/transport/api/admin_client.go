package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

type AdminClient struct {
	baseURL string
	client  *http.Client
}

func NewAdminClient(baseURL string, timeout time.Duration) *AdminClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	trimmed := strings.TrimSpace(baseURL)
	trimmed = strings.TrimRight(trimmed, "/")

	return &AdminClient{
		baseURL: trimmed,
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *AdminClient) ListUsers(ctx context.Context, limit, offset int, onlyBanned, onlyModerators bool) (domain.UserList, error) {
	if c == nil || c.client == nil {
		return domain.UserList{}, errors.New("admin client is not configured")
	}
	if c.baseURL == "" {
		return domain.UserList{}, errors.New("admin client base url is empty")
	}
	if err := ctx.Err(); err != nil {
		return domain.UserList{}, err
	}

	query := fmt.Sprintf("limit=%d&offset=%d", limit, offset)
	if onlyBanned {
		query = query + "&banned=true"
	}
	if onlyModerators {
		query = query + "&moderator=true"
	}
	url := fmt.Sprintf("%s/api/v1/admin/users?%s", c.baseURL, query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.UserList{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.UserList{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.UserList{}, c.apiError(resp)
	}

	var payload adminUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.UserList{}, err
	}

	users := make([]domain.UserSummary, 0, len(payload.Users))
	for _, user := range payload.Users {
		users = append(users, domain.UserSummary{
			UserID:      user.UserID,
			Username:    user.Username,
			Name:        user.Name,
			IsHidden:    user.IsHidden,
			IsBanned:    user.IsBanned,
			IsModerator: user.IsModerator,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		})
	}

	return domain.UserList{
		Users:  users,
		Total:  payload.Total,
		Limit:  payload.Limit,
		Offset: payload.Offset,
	}, nil
}

func (c *AdminClient) ListReportedUsers(ctx context.Context, limit, offset int) (domain.ReportedUserList, error) {
	if c == nil || c.client == nil {
		return domain.ReportedUserList{}, errors.New("admin client is not configured")
	}
	if c.baseURL == "" {
		return domain.ReportedUserList{}, errors.New("admin client base url is empty")
	}
	if err := ctx.Err(); err != nil {
		return domain.ReportedUserList{}, err
	}

	query := fmt.Sprintf("limit=%d&offset=%d", limit, offset)
	url := fmt.Sprintf("%s/api/v1/admin/reports?%s", c.baseURL, query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.ReportedUserList{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.ReportedUserList{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.ReportedUserList{}, c.apiError(resp)
	}

	var payload adminReportedUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.ReportedUserList{}, err
	}

	users := make([]domain.ReportedUserSummary, 0, len(payload.Users))
	for _, user := range payload.Users {
		users = append(users, domain.ReportedUserSummary{
			UserID:      user.UserID,
			Username:    user.Username,
			Name:        user.Name,
			IsHidden:    user.IsHidden,
			IsBanned:    user.IsBanned,
			IsModerator: user.IsModerator,
			ReportCount: user.ReportCount,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		})
	}

	return domain.ReportedUserList{
		Users:  users,
		Total:  payload.Total,
		Limit:  payload.Limit,
		Offset: payload.Offset,
	}, nil
}

func (c *AdminClient) GetUser(ctx context.Context, userID int64) (domain.UserSummary, error) {
	if userID <= 0 {
		return domain.UserSummary{}, errors.New("user id is required")
	}
	return c.getAdminUser(ctx, fmt.Sprintf("%d", userID))
}

func (c *AdminClient) GetUserByUsername(ctx context.Context, username string) (domain.UserSummary, error) {
	ref, err := normalizeAdminUsernameRef(username)
	if err != nil {
		return domain.UserSummary{}, err
	}

	return c.getAdminUser(ctx, ref)
}

func (c *AdminClient) BanUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("user id is required")
	}
	return c.postAdminAction(ctx, fmt.Sprintf("%d", userID), "ban")
}

func (c *AdminClient) BanUserByUsername(ctx context.Context, username string) error {
	ref, err := normalizeAdminUsernameRef(username)
	if err != nil {
		return err
	}

	return c.postAdminAction(ctx, ref, "ban")
}

func (c *AdminClient) UnbanUser(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("user id is required")
	}
	return c.postAdminAction(ctx, fmt.Sprintf("%d", userID), "unban")
}

func (c *AdminClient) UnbanUserByUsername(ctx context.Context, username string) error {
	ref, err := normalizeAdminUsernameRef(username)
	if err != nil {
		return err
	}

	return c.postAdminAction(ctx, ref, "unban")
}

func (c *AdminClient) MakeModerator(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("user id is required")
	}
	return c.postAdminAction(ctx, fmt.Sprintf("%d", userID), "moderator")
}

func (c *AdminClient) MakeModeratorByUsername(ctx context.Context, username string) error {
	ref, err := normalizeAdminUsernameRef(username)
	if err != nil {
		return err
	}

	return c.postAdminAction(ctx, ref, "moderator")
}

func (c *AdminClient) RemoveModerator(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("user id is required")
	}
	return c.postAdminAction(ctx, fmt.Sprintf("%d", userID), "unmoderator")
}

func (c *AdminClient) RemoveModeratorByUsername(ctx context.Context, username string) error {
	ref, err := normalizeAdminUsernameRef(username)
	if err != nil {
		return err
	}

	return c.postAdminAction(ctx, ref, "unmoderator")
}

func (c *AdminClient) getAdminUser(ctx context.Context, userRef string) (domain.UserSummary, error) {
	if c == nil || c.client == nil {
		return domain.UserSummary{}, errors.New("admin client is not configured")
	}
	if c.baseURL == "" {
		return domain.UserSummary{}, errors.New("admin client base url is empty")
	}
	if err := ctx.Err(); err != nil {
		return domain.UserSummary{}, err
	}

	normalized := strings.TrimSpace(userRef)
	if normalized == "" {
		return domain.UserSummary{}, errors.New("user reference is required")
	}

	endpoint := fmt.Sprintf("%s/api/v1/admin/users/%s", c.baseURL, url.PathEscape(normalized))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.UserSummary{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.UserSummary{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.UserSummary{}, c.apiError(resp)
	}

	var payload adminUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.UserSummary{}, err
	}

	return domain.UserSummary{
		UserID:      payload.UserID,
		Username:    payload.Username,
		Name:        payload.Name,
		IsHidden:    payload.IsHidden,
		IsBanned:    payload.IsBanned,
		IsModerator: payload.IsModerator,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}

func (c *AdminClient) postAdminAction(ctx context.Context, userRef, action string) error {
	if c == nil || c.client == nil {
		return errors.New("admin client is not configured")
	}
	if c.baseURL == "" {
		return errors.New("admin client base url is empty")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	normalized := strings.TrimSpace(userRef)
	if normalized == "" {
		return errors.New("user reference is required")
	}

	endpoint := fmt.Sprintf("%s/api/v1/admin/users/%s/%s", c.baseURL, url.PathEscape(normalized), action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	return c.apiError(resp)
}

func (c *AdminClient) apiError(resp *http.Response) error {
	message := c.extractError(resp)
	if message == "" {
		message = resp.Status
	}
	return &apiError{status: resp.StatusCode, message: message}
}

func (c *AdminClient) extractError(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		return ""
	}

	var payload errorResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}

	return strings.TrimSpace(payload.Error)
}

type adminUsersResponse struct {
	Users  []adminUserResponse `json:"users"`
	Total  int                 `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

type adminUserResponse struct {
	UserID      int64     `json:"user_id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	IsHidden    bool      `json:"is_hidden"`
	IsBanned    bool      `json:"is_banned"`
	IsModerator bool      `json:"is_moderator"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type adminReportedUsersResponse struct {
	Users  []adminReportedUserResponse `json:"users"`
	Total  int                         `json:"total"`
	Limit  int                         `json:"limit"`
	Offset int                         `json:"offset"`
}

type adminReportedUserResponse struct {
	UserID      int64     `json:"user_id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	IsHidden    bool      `json:"is_hidden"`
	IsBanned    bool      `json:"is_banned"`
	IsModerator bool      `json:"is_moderator"`
	ReportCount int       `json:"report_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func normalizeAdminUsernameRef(username string) (string, error) {
	normalized := strings.TrimSpace(username)
	if normalized == "" {
		return "", errors.New("username is required")
	}
	normalized = strings.TrimPrefix(normalized, "@")
	if normalized == "" {
		return "", errors.New("username is required")
	}

	return "@" + normalized, nil
}
