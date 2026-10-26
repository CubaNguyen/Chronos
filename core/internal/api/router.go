package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Route cơ bản để check server sống hay chết
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "Chronos Go Core",
		})
	})

	// Nhóm route API v1 (sẽ dùng sau này cho F10, F7)
	v1 := r.Group("/api/v1")
	{
		// v1.POST("/query", ...)
		// v1.GET("/graph", ...)
		_ = v1 
	}

	return r
}