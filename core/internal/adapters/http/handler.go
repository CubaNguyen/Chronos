package http

import (
	"net/http"

	"github.com/cuba3/chronos/core/internal/core/pipeline"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	pipelineService *pipeline.Service
}

func NewHandler(ps *pipeline.Service) *Handler {
	return &Handler{pipelineService: ps}
}

type ParseRequest struct {
	Path   string `json:"path" binding:"required"`
	Source string `json:"source" binding:"required"`
}

func (h *Handler) HandleParse(c *gin.Context) {
	var req ParseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	ast, err := h.pipelineService.ParseCode(c.Request.Context(), req.Path, req.Source)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Parse thành công",
		"data":    ast,
	})
}
