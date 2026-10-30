package search

import (
	"context"
	"fmt"

	"github.com/cuba3/chronos/core/internal/port"
)

type Service struct {
	aiClient   port.AI
	vectorRepo port.VectorRepository
}

func NewService(ai port.AI, repo port.VectorRepository) *Service {
	return &Service{
		aiClient:   ai,
		vectorRepo: repo,
	}
}

// Search là logic nghiệp vụ: Nhận câu hỏi -> Lấy Vector -> Tìm trong DB
func (s *Service) Search(ctx context.Context, repoName string, queryText string, limit int) ([]port.CodeEmbedding, error) {
	if repoName == "" || queryText == "" {
		return nil, fmt.Errorf("repo_name và câu hỏi không được để trống")
	}

	// 1. Gọi Python để biến câu hỏi thành Vector (F5)
	// Hàm GetEmbeddings trả về mảng 2 chiều, nên ta lấy phần tử đầu tiên [0]
	embeddings, err := s.aiClient.GetEmbeddings(ctx, []string{queryText})
	if err != nil {
		return nil, fmt.Errorf("lỗi khi tạo embedding: %w", err)
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("AI không trả về vector nào")
	}
	queryVector := embeddings[0]

	// 2. Dùng Vector tìm trong Postgres (F6)
	results, err := s.vectorRepo.SearchSimilar(ctx, repoName, queryVector, limit)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi tìm kiếm vector: %w", err)
	}

	return results, nil
}
