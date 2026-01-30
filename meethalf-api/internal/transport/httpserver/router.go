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
	searchGroup.POST("/ai", handlers.Search.AI)
	searchGroup.POST("/next", handlers.Search.Next)
	searchGroup.POST("/previous", handlers.Search.Previous)
	searchGroup.POST("/action", handlers.Search.Action)
	searchGroup.GET("/history/:user_id", handlers.Search.History)

	likesGroup := v1.Group("/likes")
	likesGroup.GET("/:user_id", handlers.Search.PendingLikes)

	adminGroup := v1.Group("/admin")
	adminGroup.GET("/users", handlers.Admin.ListUsers)
	adminGroup.GET("/users/:user_ref", handlers.Admin.GetUser)
	adminGroup.GET("/reports", handlers.Admin.ListReportedUsers)
	adminGroup.POST("/users/:user_ref/ban", handlers.Admin.BanUser)
	adminGroup.POST("/users/:user_ref/unban", handlers.Admin.UnbanUser)
	adminGroup.POST("/users/:user_ref/shadow-ban", handlers.Admin.ShadowBanUser)
	adminGroup.POST("/users/:user_ref/shadow-unban", handlers.Admin.UnshadowBanUser)
	adminGroup.POST("/users/:user_ref/hide", handlers.Admin.HideUser)
	adminGroup.POST("/users/:user_ref/show", handlers.Admin.ShowUser)
	adminGroup.POST("/users/:user_ref/moderator", handlers.Admin.MakeModerator)
	adminGroup.POST("/users/:user_ref/unmoderator", handlers.Admin.RemoveModerator)
	adminGroup.POST("/users/:user_ref/reports/clear", handlers.Admin.ClearUserReports)
	adminGroup.POST("/users/:user_ref/choices/reset", handlers.Admin.ResetChoices)

	return router
}
