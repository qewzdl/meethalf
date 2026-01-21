package httpserver

import "github.com/gin-gonic/gin"

type errorResponse struct {
	Error string `json:"error"`
}

func respondError(c *gin.Context, code int, message string) {
	c.JSON(code, errorResponse{Error: message})
}
