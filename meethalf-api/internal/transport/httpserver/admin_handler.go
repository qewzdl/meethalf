package httpserver

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"meethalf-api/internal/domain"
	"meethalf-api/internal/usecase/admin"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	uc admin.Usecase
}

func NewAdminHandler(uc admin.Usecase) *AdminHandler {
	return &AdminHandler{uc: uc}
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

func (h *AdminHandler) ListUsers(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "admin handler is not configured")
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

	onlyBanned, err := parseOptionalBoolQuery(c, "banned")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid banned flag")
		return
	}

	onlyModerators, err := parseOptionalBoolQuery(c, "moderator")
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid moderator flag")
		return
	}

	list, err := h.uc.ListUsers(c.Request.Context(), limit, offset, onlyBanned, onlyModerators)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to list users")
		return
	}

	resp := adminUsersResponse{
		Users:  make([]adminUserResponse, 0, len(list.Users)),
		Total:  list.Total,
		Limit:  list.Limit,
		Offset: list.Offset,
	}
	for _, user := range list.Users {
		resp.Users = append(resp.Users, toAdminUserResponse(user))
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AdminHandler) BanUser(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "admin handler is not configured")
		return
	}

	userID, username, ok := parseAdminUserReference(c.Param("user_ref"))
	if !ok {
		respondError(c, http.StatusBadRequest, "invalid user reference")
		return
	}

	if err := h.applyAdminBan(c, userID, username); err != nil {
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) applyAdminBan(c *gin.Context, userID int64, username string) error {
	var err error
	if username != "" {
		err = h.uc.BanUserByUsername(c.Request.Context(), username)
	} else {
		err = h.uc.BanUser(c.Request.Context(), userID)
	}
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrInvalidUserID):
			respondError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, admin.ErrInvalidUsername):
			respondError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, admin.ErrUserNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, "failed to ban user")
		}
		return err
	}

	return nil
}

func (h *AdminHandler) UnbanUser(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "admin handler is not configured")
		return
	}

	userID, username, ok := parseAdminUserReference(c.Param("user_ref"))
	if !ok {
		respondError(c, http.StatusBadRequest, "invalid user reference")
		return
	}

	if err := h.applyAdminUnban(c, userID, username); err != nil {
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) applyAdminUnban(c *gin.Context, userID int64, username string) error {
	var err error
	if username != "" {
		err = h.uc.UnbanUserByUsername(c.Request.Context(), username)
	} else {
		err = h.uc.UnbanUser(c.Request.Context(), userID)
	}
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrInvalidUserID):
			respondError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, admin.ErrInvalidUsername):
			respondError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, admin.ErrUserNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, "failed to unban user")
		}
		return err
	}

	return nil
}

func (h *AdminHandler) MakeModerator(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "admin handler is not configured")
		return
	}

	userID, username, ok := parseAdminUserReference(c.Param("user_ref"))
	if !ok {
		respondError(c, http.StatusBadRequest, "invalid user reference")
		return
	}

	if err := h.applyAdminModerator(c, userID, username); err != nil {
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) applyAdminModerator(c *gin.Context, userID int64, username string) error {
	var err error
	if username != "" {
		err = h.uc.MakeModeratorByUsername(c.Request.Context(), username)
	} else {
		err = h.uc.MakeModerator(c.Request.Context(), userID)
	}
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrInvalidUserID):
			respondError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, admin.ErrInvalidUsername):
			respondError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, admin.ErrUserNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, "failed to set moderator role")
		}
		return err
	}

	return nil
}

func (h *AdminHandler) RemoveModerator(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "admin handler is not configured")
		return
	}

	userID, username, ok := parseAdminUserReference(c.Param("user_ref"))
	if !ok {
		respondError(c, http.StatusBadRequest, "invalid user reference")
		return
	}

	if err := h.applyAdminUnmoderator(c, userID, username); err != nil {
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) applyAdminUnmoderator(c *gin.Context, userID int64, username string) error {
	var err error
	if username != "" {
		err = h.uc.RemoveModeratorByUsername(c.Request.Context(), username)
	} else {
		err = h.uc.RemoveModerator(c.Request.Context(), userID)
	}
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrInvalidUserID):
			respondError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, admin.ErrInvalidUsername):
			respondError(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, admin.ErrUserNotFound):
			respondError(c, http.StatusNotFound, err.Error())
		default:
			respondError(c, http.StatusInternalServerError, "failed to remove moderator role")
		}
		return err
	}

	return nil
}

func toAdminUserResponse(user domain.UserSummary) adminUserResponse {
	return adminUserResponse{
		UserID:      user.UserID,
		Username:    user.Username,
		Name:        user.Name,
		IsHidden:    user.IsHidden,
		IsBanned:    user.IsBanned,
		IsModerator: user.IsModerator,
		CreatedAt:   user.CreatedAt.UTC(),
		UpdatedAt:   user.UpdatedAt.UTC(),
	}
}

func parseNonNegativeIntQuery(c *gin.Context, key string) (int, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	if value < 0 {
		return 0, fmt.Errorf("%s must be non-negative", key)
	}

	return value, nil
}

func parseOptionalBoolQuery(c *gin.Context, key string) (bool, error) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return false, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, err
	}

	return value, nil
}

func parseAdminUserReference(raw string) (int64, string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, "", false
	}

	if strings.HasPrefix(raw, "@") {
		username := strings.TrimSpace(strings.TrimPrefix(raw, "@"))
		if username == "" {
			return 0, "", false
		}
		return 0, username, true
	}

	userID, err := strconv.ParseInt(raw, 10, 64)
	if err == nil && userID > 0 {
		return userID, "", true
	}

	return 0, raw, true
}
