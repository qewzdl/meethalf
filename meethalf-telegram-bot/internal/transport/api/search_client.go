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

type SearchClient struct {
	baseURL string
	client  *http.Client
}

func NewSearchClient(baseURL string, timeout time.Duration) *SearchClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	trimmed := strings.TrimSpace(baseURL)
	trimmed = strings.TrimRight(trimmed, "/")

	return &SearchClient{
		baseURL: trimmed,
		client:  &http.Client{Timeout: timeout},
	}
}

type searchStartRequest struct {
	UserID   int64         `json:"user_id"`
	Gender   domain.Gender `json:"gender"`
	Accuracy int           `json:"accuracy"`
}

type searchUserRequest struct {
	UserID int64 `json:"user_id"`
}

type searchActionRequest struct {
	UserID   int64              `json:"user_id"`
	TargetID int64              `json:"target_id"`
	Action   domain.MatchAction `json:"action"`
}

type searchCandidateResponse struct {
	Profile     profileResponse `json:"profile"`
	HasPrevious bool            `json:"has_previous"`
}

type likesResponse struct {
	Likes []profileResponse `json:"likes"`
}

type historyResponse struct {
	History []historyItemResponse `json:"history"`
	Total   int                   `json:"total"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
}

type historyItemResponse struct {
	Profile  profileResponse    `json:"profile"`
	Position int                `json:"position"`
	Action   domain.MatchAction `json:"action"`
}

type searchActionResponse struct {
	Matched bool `json:"matched"`
}

func (c *SearchClient) StartSearch(ctx context.Context, userID int64, gender domain.Gender, accuracy int) (domain.MatchCandidate, bool, error) {
	if c == nil || c.client == nil {
		return domain.MatchCandidate{}, false, errors.New("search client is not configured")
	}
	if c.baseURL == "" {
		return domain.MatchCandidate{}, false, errors.New("search client base url is empty")
	}
	if userID <= 0 {
		return domain.MatchCandidate{}, false, errors.New("user id is required")
	}
	if err := ctx.Err(); err != nil {
		return domain.MatchCandidate{}, false, err
	}

	payload, err := json.Marshal(searchStartRequest{
		UserID:   userID,
		Gender:   gender,
		Accuracy: accuracy,
	})
	if err != nil {
		return domain.MatchCandidate{}, false, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/v1/search/start",
		bytes.NewReader(payload),
	)
	if err != nil {
		return domain.MatchCandidate{}, false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.MatchCandidate{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return domain.MatchCandidate{}, false, nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.MatchCandidate{}, false, c.apiError(resp)
	}

	var payloadResp searchCandidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&payloadResp); err != nil {
		return domain.MatchCandidate{}, false, err
	}

	profile, err := profileFromResponse(payloadResp.Profile)
	if err != nil {
		return domain.MatchCandidate{}, false, err
	}

	return domain.MatchCandidate{Profile: profile, HasPrevious: payloadResp.HasPrevious}, true, nil
}

func (c *SearchClient) NextCandidate(ctx context.Context, userID int64) (domain.MatchCandidate, bool, error) {
	return c.fetchCandidate(ctx, "/api/v1/search/next", searchUserRequest{UserID: userID})
}

func (c *SearchClient) PreviousCandidate(ctx context.Context, userID int64) (domain.MatchCandidate, bool, error) {
	return c.fetchCandidate(ctx, "/api/v1/search/previous", searchUserRequest{UserID: userID})
}

func (c *SearchClient) RecordAction(ctx context.Context, userID, targetID int64, action domain.MatchAction) (domain.MatchActionResult, error) {
	if c == nil || c.client == nil {
		return domain.MatchActionResult{}, errors.New("search client is not configured")
	}
	if c.baseURL == "" {
		return domain.MatchActionResult{}, errors.New("search client base url is empty")
	}
	if userID <= 0 {
		return domain.MatchActionResult{}, errors.New("user id is required")
	}
	if targetID <= 0 {
		return domain.MatchActionResult{}, errors.New("target id is required")
	}
	if err := ctx.Err(); err != nil {
		return domain.MatchActionResult{}, err
	}

	payload, err := json.Marshal(searchActionRequest{
		UserID:   userID,
		TargetID: targetID,
		Action:   action,
	})
	if err != nil {
		return domain.MatchActionResult{}, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/api/v1/search/action",
		bytes.NewReader(payload),
	)
	if err != nil {
		return domain.MatchActionResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.MatchActionResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return domain.MatchActionResult{}, nil
	}
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		var payloadResp searchActionResponse
		if err := json.NewDecoder(resp.Body).Decode(&payloadResp); err != nil {
			if errors.Is(err, io.EOF) {
				return domain.MatchActionResult{}, nil
			}
			return domain.MatchActionResult{}, err
		}
		return domain.MatchActionResult{Matched: payloadResp.Matched}, nil
	}

	return domain.MatchActionResult{}, c.apiError(resp)
}

func (c *SearchClient) PendingLikes(ctx context.Context, userID int64) ([]domain.Profile, error) {
	if c == nil || c.client == nil {
		return nil, errors.New("search client is not configured")
	}
	if c.baseURL == "" {
		return nil, errors.New("search client base url is empty")
	}
	if userID <= 0 {
		return nil, errors.New("user id is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/likes/%d", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, c.apiError(resp)
	}

	var payload likesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	likes := make([]domain.Profile, 0, len(payload.Likes))
	for _, item := range payload.Likes {
		profile, err := profileFromResponse(item)
		if err != nil {
			return nil, err
		}
		likes = append(likes, profile)
	}

	return likes, nil
}

func (c *SearchClient) History(ctx context.Context, userID int64, limit, offset int) (domain.MatchHistoryList, error) {
	if c == nil || c.client == nil {
		return domain.MatchHistoryList{}, errors.New("search client is not configured")
	}
	if c.baseURL == "" {
		return domain.MatchHistoryList{}, errors.New("search client base url is empty")
	}
	if userID <= 0 {
		return domain.MatchHistoryList{}, errors.New("user id is required")
	}
	if err := ctx.Err(); err != nil {
		return domain.MatchHistoryList{}, err
	}

	query := fmt.Sprintf("limit=%d&offset=%d", limit, offset)
	url := fmt.Sprintf("%s/api/v1/search/history/%d?%s", c.baseURL, userID, query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return domain.MatchHistoryList{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.MatchHistoryList{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.MatchHistoryList{}, c.apiError(resp)
	}

	var payload historyResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.MatchHistoryList{}, err
	}

	items := make([]domain.MatchHistoryItem, 0, len(payload.History))
	for _, item := range payload.History {
		profile, err := profileFromResponse(item.Profile)
		if err != nil {
			return domain.MatchHistoryList{}, err
		}
		items = append(items, domain.MatchHistoryItem{
			Profile:  profile,
			Position: item.Position,
			Action:   item.Action,
		})
	}

	return domain.MatchHistoryList{
		Items:  items,
		Total:  payload.Total,
		Limit:  payload.Limit,
		Offset: payload.Offset,
	}, nil
}

func (c *SearchClient) fetchCandidate(ctx context.Context, path string, request searchUserRequest) (domain.MatchCandidate, bool, error) {
	if c == nil || c.client == nil {
		return domain.MatchCandidate{}, false, errors.New("search client is not configured")
	}
	if c.baseURL == "" {
		return domain.MatchCandidate{}, false, errors.New("search client base url is empty")
	}
	if request.UserID <= 0 {
		return domain.MatchCandidate{}, false, errors.New("user id is required")
	}
	if err := ctx.Err(); err != nil {
		return domain.MatchCandidate{}, false, err
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return domain.MatchCandidate{}, false, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+path,
		bytes.NewReader(payload),
	)
	if err != nil {
		return domain.MatchCandidate{}, false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.MatchCandidate{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return domain.MatchCandidate{}, false, nil
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return domain.MatchCandidate{}, false, c.apiError(resp)
	}

	var payloadResp searchCandidateResponse
	if err := json.NewDecoder(resp.Body).Decode(&payloadResp); err != nil {
		return domain.MatchCandidate{}, false, err
	}

	profile, err := profileFromResponse(payloadResp.Profile)
	if err != nil {
		return domain.MatchCandidate{}, false, err
	}

	return domain.MatchCandidate{Profile: profile, HasPrevious: payloadResp.HasPrevious}, true, nil
}

func (c *SearchClient) apiError(resp *http.Response) error {
	message := c.extractError(resp)
	if message == "" {
		message = resp.Status
	}
	return &apiError{status: resp.StatusCode, message: message}
}

func (c *SearchClient) extractError(resp *http.Response) string {
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

type apiError struct {
	status  int
	message string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("api error: %s", e.message)
}

func (e *apiError) StatusCode() int {
	return e.status
}

func profileFromResponse(payload profileResponse) (domain.Profile, error) {
	birthDate, err := parseBirthDate(payload.BirthDate)
	if err != nil {
		return domain.Profile{}, err
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
		IsHidden:    payload.IsHidden,
		CreatedAt:   payload.CreatedAt,
		UpdatedAt:   payload.UpdatedAt,
	}, nil
}
