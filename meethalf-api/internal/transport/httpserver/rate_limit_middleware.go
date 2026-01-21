package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"meethalf-api/internal/ratelimit"
)

func rateLimitMiddleware(limiter ratelimit.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, errorResponse{Error: "rate_limit_exceeded"})
			return
		}

		c.Next()
	}
}
