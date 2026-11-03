package query

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cuba3/chronos/core/internal/port"
)

type Service struct {
	aiClient   port.AI
	vectorRepo port.VectorRepository
	graphRepo  port.GraphRepository
}

func NewService(ai port.AI, repo port.VectorRepository, graph port.GraphRepository) *Service {
	return &Service{aiClient: ai, vectorRepo: repo, graphRepo: graph}
}

func (s *Service) GetImpact(ctx context.Context, functionName string) ([]port.GraphNode, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}
	return s.graphRepo.GetCallers(ctx, functionName)
}

// Ask là phiên bản non-streaming gom toàn bộ stream lại (dùng cho REST API thông thường và CLI)
func (s *Service) Ask(ctx context.Context, repoName string, question string) (*port.QueryResult, error) {
	stream, err := s.AskStream(ctx, repoName, question)
	if err != nil {
		return nil, err
	}

	var finalResult *port.QueryResult
	var fullAnswer strings.Builder

	for res := range stream {
		if res == nil {
			continue
		}
		fullAnswer.WriteString(res.Answer)
		finalResult = res
	}

	if finalResult == nil {
		return nil, fmt.Errorf("không nhận được phản hồi từ AI")
	}

	finalResult.Answer = fullAnswer.String()
	return finalResult, nil
}

func (s *Service) AskStream(ctx context.Context, repoName string, question string) (<-chan *port.QueryResult, error) {
	if question == "" {
		return nil, fmt.Errorf("câu hỏi không được để trống")
	}

	// 1. Get Embedding
	embeddings, err := s.aiClient.GetEmbeddings(ctx, []string{question})
	if err != nil {
		return nil, fmt.Errorf("lỗi embedding: %w", err)
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("AI không trả về vector")
	}

	// 2. Search DB - lấy top 10 chunks để không bị lọt ngữ cảnh
	dbChunks, err := s.vectorRepo.SearchSimilar(ctx, repoName, embeddings[0], 10)
	if err != nil {
		return nil, fmt.Errorf("lỗi search db: %w", err)
	}
	slog.Info("Tìm thấy code tương đồng", "count", len(dbChunks))

	// 3. Map to Port Chunks
	contexts := make([]port.Chunk, 0, len(dbChunks))
	for _, c := range dbChunks {
		meta := make(map[string]string)
		for k, v := range c.Metadata {
			meta[k] = fmt.Sprintf("%v", v)
		}
		meta["name"] = c.FunctionName
		contexts = append(contexts, port.Chunk{
			ID: c.ID, Text: c.Content, SourceType: "function",
			SourcePath: c.FilePath, Metadata: meta,
		})
	}

	// 4. Call Stream
	return s.aiClient.QueryStream(ctx, question, 5, 0.3, contexts)
}
