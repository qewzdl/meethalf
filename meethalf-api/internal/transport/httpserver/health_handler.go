package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"meethalf-api/internal/domain"
	"meethalf-api/internal/usecase/health"
)

type HealthHandler struct {
	uc health.Usecase
}

func NewHealthHandler(uc health.Usecase) *HealthHandler {
	return &HealthHandler{uc: uc}
}

func (h *HealthHandler) Liveness(c *gin.Context) {
	status := h.uc.Liveness(c.Request.Context())
	c.JSON(http.StatusOK, status)
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	status := h.uc.Readiness(c.Request.Context())
	c.JSON(healthHTTPStatus(status.Status), status)
}

func (h *HealthHandler) Status(c *gin.Context) {
	status := h.uc.Status(c.Request.Context())
	c.JSON(healthHTTPStatus(status.Status), status)
}

func healthHTTPStatus(status string) int {
	if status == domain.HealthStatusOK {
		return http.StatusOK
	}

	return http.StatusServiceUnavailable
}
