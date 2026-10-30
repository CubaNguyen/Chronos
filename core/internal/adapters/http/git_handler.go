package http

import (
	"context"
	"net/http"

	"github.com/cuba3/chronos/core/internal/core/ingest"
	"github.com/gin-gonic/gin"
)

type GitHandler struct {
	gitIngestService *ingest.GitIngestService
}

func NewGitHandler(gis *ingest.GitIngestService) *GitHandler {
	return &GitHandler{gitIngestService: gis}
}

type IngestDirRequest struct {
	RepoName string `json:"repo_name" binding:"required"`
	DirPath  string `json:"dir_path" binding:"required"`
}

func (h *GitHandler) HandleIngestDir(c *gin.Context) {
	var req IngestDirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Chạy ngầm hàm quét thư mục để API không bị block
	go func() {
		ctx := context.WithoutCancel(c.Request.Context())
		err := h.gitIngestService.IngestDirectory(ctx, req.RepoName, req.DirPath)
		if err != nil {
			// Log lỗi ra (không thể trả về HTTP vì API đã trả về 202 trước đó)
			println("Lỗi Ingest Directory:", err.Error())
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{"message": "Đã bắt đầu quét thư mục ngầm"})
}
