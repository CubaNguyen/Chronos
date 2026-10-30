package pipeline

import (
	"context"
	"fmt"

	"github.com/cuba3/chronos/core/internal/port"
)

type Service struct {
	parser port.Parser
}

func NewService(parser port.Parser) *Service {
	return &Service{
		parser: parser,
	}
}

func (s *Service) ParseCode(ctx context.Context, path string, source string) (*port.Ast, error) {
	if source == "" {
		return nil, fmt.Errorf("source code không được để trống")
	}

	ast, err := s.parser.ParseFile(ctx, path, source)
	if err != nil {
		return nil, fmt.Errorf("lỗi từ parser: %w", err)
	}

	return ast, nil
}
