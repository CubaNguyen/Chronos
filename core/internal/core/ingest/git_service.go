package ingest

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/cuba3/chronos/core/internal/port"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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

func (s *GitIngestService) IngestGitHistory(ctx context.Context, repoName string, repoPath string) error {
	slog.Info("Bắt đầu đọc lịch sử Git", "repo", repoName, "path", repoPath)

	// 1. Open the Git repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("lỗi mở git repo: %w", err)
	}

	// 2. Get the HEAD (latest commit)
	ref, err := repo.Head()
	if err != nil {
		return fmt.Errorf("lỗi lấy HEAD: %w", err)
	}
	commitHash := ref.Hash().String() // 🚀 Get the real Git Hash
	slog.Info("Đang xử lý commit", "hash", commitHash)

	// 3. Get the commit object
	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return fmt.Errorf("lỗi lấy commit: %w", err)
	}

	// 4. Get the file tree of this commit
	tree, err := commit.Tree()
	if err != nil {
		return fmt.Errorf("lỗi lấy tree: %w", err)
	}

	count := 0
	// 5. Iterate through all files in the commit tree
	err = tree.Files().ForEach(func(file *object.File) error {
		// Only ingest .go files (for now)
		if !strings.HasSuffix(file.Name, ".go") {
			return nil
		}

		// Read the file content from Git, NOT from the disk
		reader, err := file.Reader()
		if err != nil {
			return err
		}
		defer reader.Close()

		contentBytes, err := io.ReadAll(reader)
		if err != nil {
			return err
		}

		// Create the Job with the REAL commit hash
		job := port.Job{
			RepoName:   repoName,
			FilePath:   file.Name,
			Language:   "go",
			SourceCode: string(contentBytes),
			CommitHash: commitHash, // 🚀 Attach real hash
		}

		// Push to NATS
		if err := s.queue.PublishJob(ctx, job); err != nil {
			slog.Error("Lỗi bắn job vào NATS", "path", file.Name, "error", err)
		} else {
			count++
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("lỗi duyệt tree: %w", err)
	}

	slog.Info("Đã bắn toàn bộ job từ Git vào NATS!", "total_files", count)
	return nil
}
