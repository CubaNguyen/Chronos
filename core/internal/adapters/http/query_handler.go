package http

import (
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
