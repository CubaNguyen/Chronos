package http

import (
	"io"
	"net/http"

	"github.com/cuba3/chronos/core/internal/core/query"
	"github.com/gin-gonic/gin"
)

type QueryHandler struct {
	queryService *query.Service
}

func NewQueryHandler(qs *query.Service) *QueryHandler {
	return &QueryHandler{queryService: qs}
}

type AskRequest struct {
	RepoName string `json:"repo_name"`
	Question string `json:"question" binding:"required"`
}

func (h *QueryHandler) HandleAsk(c *gin.Context) {
	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	// Gọi xuống tầng Core (kèm repoName nếu có)
	result, err := h.queryService.Ask(c.Request.Context(), req.RepoName, req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Trả về kết quả
	c.JSON(http.StatusOK, gin.H{
		"message": "Truy vấn thành công",
		"data":    result,
	})
}

func (h *QueryHandler) HandleAskStream(c *gin.Context) {
	var req AskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Call Core Service to get the channel
	resultChan, err := h.queryService.AskStream(c.Request.Context(), req.RepoName, req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Set SSE Headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 3. Stream the data
	c.Stream(func(w io.Writer) bool {
		// Read from the channel
		result, ok := <-resultChan
		if !ok {
			// Channel is closed -> Stream finished
			c.SSEvent("message", gin.H{
				"answer": "",
				"done":   true,
			})
			return false
		}

		// Write data in SSE format: "data: {text}\n\n"
		c.SSEvent("message", gin.H{
			"answer": result.Answer,
			"done":   false,
		})

		return true // Keep the connection open
	})
}

func (h *QueryHandler) HandleImpact(c *gin.Context) {
	funcName := c.Query("function") // Get from URL: ?function=AskCmd
	if funcName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "function query parameter is required"})
		return
	}

	impactedNodes, err := h.queryService.GetImpact(c.Request.Context(), funcName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"target_function":    funcName,
		"blast_radius_count": len(impactedNodes),
		"affected_functions": impactedNodes,
	})
}
