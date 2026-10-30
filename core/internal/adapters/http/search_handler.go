package http

import (
	"net/http"

	"github.com/cuba3/chronos/core/internal/core/search"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	searchService *search.Service
}

func NewSearchHandler(ss *search.Service) *SearchHandler {
	return &SearchHandler{searchService: ss}
}

type SearchRequest struct {
	RepoName string `json:"repo_name" binding:"required"`
	Query    string `json:"query" binding:"required"`
}

func (h *SearchHandler) HandleSearch(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	results, err := h.searchService.Search(c.Request.Context(), req.RepoName, req.Query, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tìm kiếm thành công",
		"data":    results,
	})
}