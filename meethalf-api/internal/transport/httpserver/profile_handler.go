package httpserver

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"meethalf-api/internal/domain"
	"meethalf-api/internal/usecase/profile"
)

const birthDateLayout = "2006-01-02"

type ProfileHandler struct {
	uc profile.Usecase
}

func NewProfileHandler(uc profile.Usecase) *ProfileHandler {
	return &ProfileHandler{uc: uc}
}

type profileRequest struct {
	UserID      int64    `json:"user_id"`
	Name        string   `json:"name"`
	Gender      string   `json:"gender"`
	BirthDate   string   `json:"birth_date"`
	Country     string   `json:"country"`
	City        string   `json:"city"`
	Description string   `json:"description"`
	EmojiCode   string   `json:"emoji_code"`
	Photos      []string `json:"photos"`
}

type profileResponse struct {
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Gender      string    `json:"gender"`
	BirthDate   string    `json:"birth_date"`
	Age         int       `json:"age"`
	Country     string    `json:"country"`
	City        string    `json:"city"`
	Description string    `json:"description"`
	EmojiCode   string    `json:"emoji_code"`
	Photos      []string  `json:"photos"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *ProfileHandler) Upsert(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "profile handler is not configured")
		return
	}

	var req profileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	birthDate, err := parseBirthDate(req.BirthDate)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid birth date")
		return
	}

	stored, err := h.uc.Upsert(c.Request.Context(), domain.Profile{
		UserID:      req.UserID,
		Name:        req.Name,
		Gender:      domain.Gender(req.Gender),
		BirthDate:   birthDate,
		Country:     domain.Country(req.Country),
		City:        req.City,
		Description: req.Description,
		EmojiCode:   domain.ProfileEmojiCode(req.EmojiCode),
		Photos:      req.Photos,
	})
	if err != nil {
		code, message := profileHTTPError(err)
		respondError(c, code, message)
		return
	}

	c.JSON(http.StatusOK, profileResponse{
		UserID:      stored.UserID,
		Name:        stored.Name,
		Gender:      string(stored.Gender),
		BirthDate:   formatBirthDate(stored.BirthDate),
		Age:         stored.Age,
		Country:     string(stored.Country),
		City:        stored.City,
		Description: stored.Description,
		EmojiCode:   string(stored.EmojiCode),
		Photos:      stored.Photos,
		CreatedAt:   stored.CreatedAt,
		UpdatedAt:   stored.UpdatedAt,
	})
}

func (h *ProfileHandler) Get(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "profile handler is not configured")
		return
	}

	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		respondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	stored, err := h.uc.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		code, message := profileHTTPError(err)
		respondError(c, code, message)
		return
	}

	c.JSON(http.StatusOK, profileResponse{
		UserID:      stored.UserID,
		Name:        stored.Name,
		Gender:      string(stored.Gender),
		BirthDate:   formatBirthDate(stored.BirthDate),
		Age:         stored.Age,
		Country:     string(stored.Country),
		City:        stored.City,
		Description: stored.Description,
		EmojiCode:   string(stored.EmojiCode),
		Photos:      stored.Photos,
		CreatedAt:   stored.CreatedAt,
		UpdatedAt:   stored.UpdatedAt,
	})
}

func (h *ProfileHandler) Delete(c *gin.Context) {
	if h == nil || h.uc == nil {
		respondError(c, http.StatusInternalServerError, "profile handler is not configured")
		return
	}

	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		respondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.uc.DeleteByUserID(c.Request.Context(), userID); err != nil {
		code, message := profileHTTPError(err)
		respondError(c, code, message)
		return
	}

	c.Status(http.StatusNoContent)
}

func profileHTTPError(err error) (int, string) {
	switch {
	case errors.Is(err, profile.ErrInvalidUserID),
		errors.Is(err, profile.ErrInvalidName),
		errors.Is(err, profile.ErrInvalidGender),
		errors.Is(err, profile.ErrInvalidBirthDate),
		errors.Is(err, profile.ErrInvalidCountry),
		errors.Is(err, profile.ErrInvalidCity),
		errors.Is(err, profile.ErrInvalidDescription),
		errors.Is(err, profile.ErrInvalidEmojiCode),
		errors.Is(err, profile.ErrInvalidPhotos):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, profile.ErrProfileNotFound):
		return http.StatusNotFound, err.Error()
	default:
		return http.StatusInternalServerError, "internal error"
	}
}

func parseBirthDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, profile.ErrInvalidBirthDate
	}

	parsed, err := time.Parse(birthDateLayout, value)
	if err != nil {
		return time.Time{}, profile.ErrInvalidBirthDate
	}

	return parsed.UTC(), nil
}

func formatBirthDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(birthDateLayout)
}
