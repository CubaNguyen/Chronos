package http

import (
	"github.com/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func SetupRouter(
	pipelineHandler *Handler,
	queryHandler *QueryHandler,
	searchHandler *SearchHandler,
	ingestHandler *IngestHandler,
	gitHandler *GitHandler,

) *gin.Engine {
	r := gin.Default()

	// Kích hoạt CORS cho toàn bộ API để Frontend gọi mượt mà không bị chặn
	r.Use(corsMiddleware())

	v1 := r.Group("/api/v1")
	{
		v1.POST("/parse", pipelineHandler.HandleParse)
		v1.POST("/ask", queryHandler.HandleAsk)
		v1.POST("/ask-stream", queryHandler.HandleAskStream)
		v1.POST("/ask/stream", queryHandler.HandleAskStream)
		v1.POST("/search", searchHandler.HandleSearch)
		v1.POST("/ingest", ingestHandler.HandleIngest)
		v1.POST("/ingest-dir", gitHandler.HandleIngestDir)
		v1.GET("/impact", queryHandler.HandleImpact)

	}

	return r
}
