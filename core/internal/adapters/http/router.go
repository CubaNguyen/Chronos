package http

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	pipelineHandler *Handler,
	queryHandler *QueryHandler,
	searchHandler *SearchHandler,
	ingestHandler *IngestHandler,
	gitHandler *GitHandler,

) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/parse", pipelineHandler.HandleParse)
		v1.POST("/ask", queryHandler.HandleAsk)
		v1.POST("/search", searchHandler.HandleSearch)
		v1.POST("/ingest", ingestHandler.HandleIngest)
		v1.POST("/ingest-dir", gitHandler.HandleIngestDir)

	}

	return r
}
