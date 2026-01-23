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

	profileGroup := v1.Group("/profiles")
	profileGroup.GET("/:user_id", handlers.Profile.Get)
	profileGroup.DELETE("/:user_id", handlers.Profile.Delete)
	profileGroup.POST("", handlers.Profile.Upsert)
	profileGroup.PATCH("/:user_id/visibility", handlers.Profile.UpdateVisibility)

	searchGroup := v1.Group("/search")
	searchGroup.POST("/start", handlers.Search.Start)
	searchGroup.POST("/next", handlers.Search.Next)
	searchGroup.POST("/previous", handlers.Search.Previous)
	searchGroup.POST("/action", handlers.Search.Action)

	likesGroup := v1.Group("/likes")
	likesGroup.GET("/:user_id", handlers.Search.PendingLikes)

	return router
}
