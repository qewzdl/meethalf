package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"meethalf-telegram-bot/internal/domain"
)

type ProfileClient struct {
	baseURL string
	client  *http.Client
}

const birthDateLayout = "2006-01-02"

func NewProfileClient(baseURL string, timeout time.Duration) *ProfileClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	trimmed := strings.TrimSpace(baseURL)
	trimmed = strings.TrimRight(trimmed, "/")

	return &ProfileClient{
		baseURL: trimmed,
		client:  &http.Client{Timeout: timeout},
	}
}

func (c *ProfileClient) CreateProfile(ctx context.Context, profile domain.Profile) error {
	if c == nil || c.client == nil {
		return errors.New("profile client is not configured")
	}

	if c.baseURL == "" {
		return errors.New("profile client base url is empty")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	payload, err := json.Marshal(profileRequest{
		UserID:      profile.UserID,
		Name:        profile.Name,
		Gender:      profile.Gender,
		BirthDate:   formatBirthDate(profile.BirthDate),
		Country:     profile.Country,
		City:        profile.City,
		Description: profile.Description,
		EmojiCode:   profile.EmojiCode,
		Photos:      profile.Photos,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/v1/profiles",
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	message := c.extractError(resp)
	if message == "" {
		message = resp.Status
	}
	return fmt.Errorf("profile api error: %s", message)
}

func (c *ProfileClient) GetProfile(ctx context.Context, userID int64) (domain.Profile, bool, error) {
	if c == nil || c.client == nil {
		return domain.Profile{}, false, errors.New("profile client is not configured")
	}

	if c.baseURL == "" {
		return domain.Profile{}, false, errors.New("profile client base url is empty")
	}

	if userID <= 0 {
		return domain.Profile{}, false, errors.New("user id is required")
	}

	if err := ctx.Err(); err != nil {
		return domain.Profile{}, false, err
	}

	url := fmt.Sprintf("%s/api/v1/profiles/%d", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.Profile{}, false, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.Profile{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return domain.Profile{}, false, nil
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message := c.extractError(resp)
		if message == "" {
			message = resp.Status
		}
		return domain.Profile{}, false, fmt.Errorf("profile api error: %s", message)
	}

	var payload profileResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.Profile{}, false, err
	}

	birthDate, err := parseBirthDate(payload.BirthDate)
	if err != nil {
		return domain.Profile{}, false, err
	}

	return domain.Profile{
		UserID:      payload.UserID,
		Name:        payload.Name,
		Gender:      payload.Gender,
		BirthDate:   birthDate,
		Age:         payload.Age,
		Country:     payload.Country,
		City:        payload.City,
		Description: payload.Description,
		EmojiCode:   payload.EmojiCode,
		Photos:      payload.Photos,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, true, nil
}

func (c *ProfileClient) DeleteProfile(ctx context.Context, userID int64) (bool, error) {
	if c == nil || c.client == nil {
		return false, errors.New("profile client is not configured")
	}

	if c.baseURL == "" {
		return false, errors.New("profile client base url is empty")
	}

	if userID <= 0 {
		return false, errors.New("user id is required")
	}

	if err := ctx.Err(); err != nil {
		return false, err
	}

	url := fmt.Sprintf("%s/api/v1/profiles/%d", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return true, nil
	}

	message := c.extractError(resp)
	if message == "" {
		message = resp.Status
	}
	return false, fmt.Errorf("profile api error: %s", message)
}

func (c *ProfileClient) extractError(resp *http.Response) string {
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

type profileRequest struct {
	UserID      int64                   `json:"user_id"`
	Name        string                  `json:"name"`
	Gender      domain.Gender           `json:"gender"`
	BirthDate   string                  `json:"birth_date"`
	Country     domain.Country          `json:"country"`
	City        string                  `json:"city"`
	Description string                  `json:"description"`
	EmojiCode   domain.ProfileEmojiCode `json:"emoji_code"`
	Photos      []string                `json:"photos"`
}

type profileResponse struct {
	UserID      int64                   `json:"user_id"`
	Name        string                  `json:"name"`
	Gender      domain.Gender           `json:"gender"`
	BirthDate   string                  `json:"birth_date"`
	Age         int                     `json:"age"`
	Country     domain.Country          `json:"country"`
	City        string                  `json:"city"`
	Description string                  `json:"description"`
	EmojiCode   domain.ProfileEmojiCode `json:"emoji_code"`
	Photos      []string                `json:"photos"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func formatBirthDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(birthDateLayout)
}

func parseBirthDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	parsed, err := time.Parse(birthDateLayout, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid birth date format: %w", err)
	}

	return parsed.UTC(), nil
}
