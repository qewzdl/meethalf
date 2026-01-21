package httpserver

import (
	"github.com/gin-gonic/gin"

	"meethalf-api/internal/ratelimit"
)

func NewRouter(handlers *Handlers, limiter ratelimit.Limiter) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	if limiter != nil {
		router.Use(rateLimitMiddleware(limiter))
	}

	v1 := router.Group("/api/v1")
	healthGroup := v1.Group("/health")
	healthGroup.GET("", handlers.Health.Status)
	healthGroup.GET("/liveness", handlers.Health.Liveness)
	healthGroup.GET("/readiness", handlers.Health.Readiness)

	return router
}
