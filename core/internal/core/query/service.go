package query

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cuba3/chronos/core/internal/port"
)

type Service struct {
	aiClient   port.AI
	vectorRepo port.VectorRepository
}

// DI: Bơm Adapter AI và VectorRepo (Postgres) vào Core
func NewService(ai port.AI, repo port.VectorRepository) *Service {
	return &Service{
		aiClient:   ai,
		vectorRepo: repo,
	}
}

// Ask là hàm xử lý logic nghiệp vụ cho F7 (RAG tập trung dữ liệu tại Postgres)
func (s *Service) Ask(ctx context.Context, repoName string, question string) (*port.QueryResult, error) {
	if question == "" {
		return nil, fmt.Errorf("câu hỏi không được để trống")
	}

	// 1. Gọi Python để lấy vector embedding của câu hỏi
	embeddings, err := s.aiClient.GetEmbeddings(ctx, []string{question})
	if err != nil {
		return nil, fmt.Errorf("lỗi khi tạo embedding câu hỏi: %w", err)
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("không nhận được vector từ AI")
	}
	questionVector := embeddings[0]

	// 2. Tìm kiếm các đoạn code tương đồng nhất trong Postgres
	dbChunks, err := s.vectorRepo.SearchSimilar(ctx, repoName, questionVector, 5)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi truy vấn vector từ Postgres: %w", err)
	}

	slog.Info("Tìm thấy code tương đồng trong Postgres cho RAG", "count", len(dbChunks))

	// 3. Chuyển đổi sang port.Chunk để gửi sang Python LLM
	contexts := make([]port.Chunk, 0, len(dbChunks))
	for _, c := range dbChunks {
		meta := make(map[string]string)
		for k, v := range c.Metadata {
			meta[k] = fmt.Sprintf("%v", v)
		}
		meta["name"] = c.FunctionName

		contexts = append(contexts, port.Chunk{
			ID:         c.ID,
			Text:       c.Content,
			SourceType: "function",
			SourcePath: c.FilePath,
			Metadata:   meta,
		})
	}

	// 4. Gọi sang Python AI: Đưa câu hỏi + 5 đoạn code từ Postgres để Gemini trả lời!
	result, err := s.aiClient.Query(ctx, question, 5, 0.3, contexts)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi truy vấn AI: %w", err)
	}

	return result, nil
}
