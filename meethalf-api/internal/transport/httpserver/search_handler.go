package httpserver

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"meethalf-api/internal/domain"
	"meethalf-api/internal/usecase/matching"
)

type SearchHandler struct {
	uc matching.Usecase
}

func NewSearchHandler(uc matching.Usecase) *SearchHandler {
	return &SearchHandler{uc: uc}
}

type searchStartRequest struct {
	UserID   int64  `json:"user_id"`
	Gender   string `json:"gender"`
	Accuracy int    `json:"accuracy"`
}

type searchUserRequest struct {
	UserID int64 `json:"user_id"`
}

type searchAIRequest struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

type searchActionRequest struct {
	UserID   int64  `json:"user_id"`
	TargetID int64  `json:"target_id"`
	Action   string `json:"action"`
}

type searchCandidateResponse struct {
	Profile     profileResponse `json:"profile"`
	HasPrevious bool            `json:"has_previous"`
}

type likesResponse struct {
	Likes []profileResponse `json:"likes"`
}

type likesListResponse struct {
	Likes  []profileResponse `json:"likes"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

type historyResponse struct {
	History []historyItemResponse `json:"history"`
	Total   int                   `json:"total"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
}

type historyItemResponse struct {
	Profile  profileResponse `json:"profile"`
	Position int             `json:"position"`
	Action   string          `json:"action"`
}

type searchActionResponse struct {
	Matched bool `json:"matched"`
}

func (h *SearchHandler) Start(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "search handler is not configured")
		return
	}

	var req searchStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	candidate, err := h.uc.Start(
		c.Request.Context(),
		req.UserID,
		domain.Gender(req.Gender),
		req.Accuracy,
	)
	if err != nil {
		if errors.Is(err, matching.ErrNoCandidates) {
			c.Status(http.StatusNoContent)
			return
		}
		code, message := searchHTTPError(err)
		respondError(c, code, message)
		return
	}

	c.JSON(http.StatusOK, searchCandidateResponse{
		Profile:     toProfileResponse(candidate.Profile),
		HasPrevious: candidate.HasPrevious,
	})
}

func (h *SearchHandler) Next(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "search handler is not configured")
		return
	}

	var req searchUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	candidate, err := h.uc.Next(c.Request.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, matching.ErrNoCandidates) {
			c.Status(http.StatusNoContent)
			return
		}
		code, message := searchHTTPError(err)
		respondError(c, code, message)
		return
	}

	c.JSON(http.StatusOK, searchCandidateResponse{
		Profile:     toProfileResponse(candidate.Profile),
		HasPrevious: candidate.HasPrevious,
	})
}

func (h *SearchHandler) AI(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "search handler is not configured")
		return
	}

	var req searchAIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	candidate, err := h.uc.SearchAI(c.Request.Context(), req.UserID, req.Message)
	if err != nil {
		if errors.Is(err, matching.ErrNoCandidates) {
			c.Status(http.StatusNoContent)
			return
		}
		code, message := searchHTTPError(err)
		respondError(c, code, message)
		return
	}

	c.JSON(http.StatusOK, searchCandidateResponse{
		Profile:     toProfileResponse(candidate.Profile),
		HasPrevious: candidate.HasPrevious,
	})
}

func (h *SearchHandler) Previous(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "search handler is not configured")
		return
	}

	var req searchUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	candidate, err := h.uc.Previous(c.Request.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, matching.ErrNoPrevious) {
			c.Status(http.StatusNoContent)
			return
		}
		code, message := searchHTTPError(err)
		respondError(c, code, message)
		return
	}

	c.JSON(http.StatusOK, searchCandidateResponse{
		Profile:     toProfileResponse(candidate.Profile),
		HasPrevious: candidate.HasPrevious,
	})
}

func (h *SearchHandler) Action(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "search handler is not configured")
		return
	}

	var req searchActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.uc.RecordAction(
		c.Request.Context(),
		req.UserID,
		req.TargetID,
		domain.MatchAction(req.Action),
	)
	if err != nil {
		code, message := searchHTTPError(err)
		respondError(c, code, message)
		return
	}

	c.JSON(http.StatusOK, searchActionResponse{Matched: result.Matched})
}

func (h *SearchHandler) PendingLikes(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "search handler is not configured")
		return
	}

	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		respondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	likes, err := h.uc.PendingLikes(c.Request.Context(), userID)
	if err != nil {
		code, message := searchHTTPError(err)
		respondError(c, code, message)
		return
	}

	resp := likesResponse{Likes: make([]profileResponse, 0, len(likes))}
	for _, like := range likes {
		resp.Likes = append(resp.Likes, toProfileResponse(like))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *SearchHandler) ReceivedLikes(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "search handler is not configured")
		return
	}

	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		respondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	limit, err := parseNonNegativeIntQuery(c, "limit")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid limit")
		return
	}

	offset, err := parseNonNegativeIntQuery(c, "offset")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid offset")
		return
	}

	list, err := h.uc.ReceivedLikes(c.Request.Context(), userID, limit, offset)
	if err != nil {
		code, message := searchHTTPError(err)
		respondError(c, code, message)
		return
	}

	resp := likesListResponse{
		Likes:  make([]profileResponse, 0, len(list.Items)),
		Total:  list.Total,
		Limit:  list.Limit,
		Offset: list.Offset,
	}
	for _, like := range list.Items {
		resp.Likes = append(resp.Likes, toProfileResponse(like))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *SearchHandler) History(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "search handler is not configured")
		return
	}

	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		respondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	limit, err := parseNonNegativeIntQuery(c, "limit")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid limit")
		return
	}

	offset, err := parseNonNegativeIntQuery(c, "offset")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid offset")
		return
	}

	list, err := h.uc.History(c.Request.Context(), userID, limit, offset)
	if err != nil {
		code, message := searchHTTPError(err)
		respondError(c, code, message)
		return
	}

	resp := historyResponse{
		History: make([]historyItemResponse, 0, len(list.Items)),
		Total:   list.Total,
		Limit:   list.Limit,
		Offset:  list.Offset,
	}
	for _, item := range list.Items {
		resp.History = append(resp.History, historyItemResponse{
			Profile:  toProfileResponse(item.Profile),
			Position: item.Position,
			Action:   string(item.Action),
		})
	}

	c.JSON(http.StatusOK, resp)
}

func searchHTTPError(err error) (int, string) {
	switch {
	case errors.Is(err, matching.ErrInvalidUserID),
		errors.Is(err, matching.ErrInvalidGender),
		errors.Is(err, matching.ErrInvalidAccuracy),
		errors.Is(err, matching.ErrInvalidAction),
		errors.Is(err, matching.ErrInvalidQuery),
		errors.Is(err, matching.ErrInvalidTargetID),
		errors.Is(err, matching.ErrInvalidSelfMatch):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, matching.ErrProfileNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, matching.ErrSessionNotFound):
		return http.StatusConflict, err.Error()
	case errors.Is(err, matching.ErrUserBanned):
		return http.StatusForbidden, err.Error()
	case errors.Is(err, matching.ErrAIUnavailable):
		return http.StatusServiceUnavailable, err.Error()
	default:
		return http.StatusInternalServerError, "internal error"
	}
}

func toProfileResponse(profile domain.Profile) profileResponse {
	return profileResponse{
		UserID:                    profile.UserID,
		Name:                      profile.Name,
		Gender:                    string(profile.Gender),
		BirthDate:                 formatBirthDate(profile.BirthDate),
		Age:                       profile.Age,
		Country:                   string(profile.Country),
		City:                      profile.City,
		Description:               profile.Description,
		EmojiCode:                 string(profile.EmojiCode),
		Photos:                    profile.Photos,
		IsHidden:                  profile.IsHidden,
		LikesNotificationsEnabled: profile.LikesNotificationsEnabled,
		CreatedAt:                 profile.CreatedAt,
		UpdatedAt:                 profile.UpdatedAt,
	}
}
