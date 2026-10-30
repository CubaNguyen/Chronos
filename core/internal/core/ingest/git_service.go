package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/cuba3/chronos/core/internal/port"
)

// GitIngestService chịu trách nhiệm quét code và bắn Job vào Queue
type GitIngestService struct {
	queue port.Queue
}

func NewGitIngestService(q port.Queue) *GitIngestService {
	return &GitIngestService{queue: q}
}

// IngestDirectory quét toàn bộ thư mục và bắn job vào NATS
func (s *GitIngestService) IngestDirectory(ctx context.Context, repoName string, dirPath string) error {
	slog.Info("Bắt đầu quét thư mục", "repo", repoName, "dir", dirPath)

	count := 0
	// filepath.Walk là hàm chuẩn của Go để duyệt đệ quy thư mục
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Bỏ qua thư mục
		if info.IsDir() {
			return nil
		}

		// Chỉ lấy file .go (Sau này bạn có thể mở rộng ra .py, .rs)
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Đọc nội dung file
		content, err := os.ReadFile(path)
		if err != nil {
			slog.Error("Không đọc được file", "path", path, "error", err)
			return nil // Bỏ qua file lỗi, chạy tiếp file khác
		}

		// Tạo Job
		job := port.Job{
			RepoName:   repoName,
			FilePath:   path,
			Language:   "go",
			SourceCode: string(content),
		}

		// Bắn vào NATS
		if err := s.queue.PublishJob(ctx, job); err != nil {
			slog.Error("Lỗi bắn job vào NATS", "path", path, "error", err)
		} else {
			count++
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("lỗi khi quét thư mục: %w", err)
	}

	slog.Info("Đã bắn toàn bộ job vào NATS!", "total_files", count)
	return nil
}
