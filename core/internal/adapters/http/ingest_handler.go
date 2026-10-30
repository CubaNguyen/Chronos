package http

import (
	"net/http"

	"github.com/cuba3/chronos/core/internal/core/ingest"
	"github.com/cuba3/chronos/core/internal/port" // 🚨 Thêm import này
	"github.com/gin-gonic/gin"
)

type IngestHandler struct {
	ingestService *ingest.Service
	queue         port.Queue
}

// Hàm khởi tạo phải nhận cả 2
func NewIngestHandler(is *ingest.Service, q port.Queue) *IngestHandler {
	return &IngestHandler{
		ingestService: is,
		queue:         q,
	}
}

type IngestRequest struct {
	RepoName string `json:"repo_name" binding:"required"`
	FilePath string `json:"file_path" binding:"required"`
	Language string `json:"language" binding:"required"`
	Source   string `json:"source" binding:"required"`
}

func (h *IngestHandler) HandleIngest(c *gin.Context) {
	var req IngestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu không hợp lệ: " + err.Error()})
		return
	}

	job := port.Job{
		RepoName:   req.RepoName,
		FilePath:   req.FilePath,
		Language:   req.Language,
		SourceCode: req.Source,
	}

	if err := h.queue.PublishJob(c.Request.Context(), job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Job đã được đưa vào hàng đợi"})
}
