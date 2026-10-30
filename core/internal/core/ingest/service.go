package ingest

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cuba3/chronos/core/internal/port"
)

type Service struct {
	parser     port.Parser
	aiClient   port.AI
	vectorRepo port.VectorRepository
}

func NewService(parser port.Parser, ai port.AI, repo port.VectorRepository) *Service {
	return &Service{
		parser:     parser,
		aiClient:   ai,
		vectorRepo: repo,
	}
}

// ProcessFile thực hiện quy trình: Parse AST -> Cắt Code -> Embedding -> Lưu Postgres
func (s *Service) ProcessFile(ctx context.Context, repoName, filePath, language, sourceCode, commitHash string) error {
	slog.Info("Worker bắt đầu xử lý file", "repo", repoName, "file", filePath)

	// 1. Gọi Parser (Rust) qua gRPC để lấy AST
	ast, err := s.parser.ParseFile(ctx, filePath, sourceCode)
	if err != nil {
		return fmt.Errorf("lỗi parse file: %w", err)
	}

	slog.Info("Parse AST thành công", "file", filePath, "functions", len(ast.Functions))

	// 2. Trích xuất code từ AST
	var textsToEmbed []string
	var chunks []port.CodeEmbedding
	lines := strings.Split(sourceCode, "\n")

	// Tạm thời chúng ta chỉ lấy các Hàm (Functions) để làm MVP
	for _, fn := range ast.Functions {
		// Cắt code dựa vào LineStart và LineEnd
		codeStr := extractCodeByLines(lines, fn.LineStart, fn.LineEnd)
		if codeStr == "" {
			continue // Bỏ qua nếu không cắt được
		}

		// Chuẩn bị metadata
		meta := map[string]interface{}{
			"line_start": fn.LineStart,
			"line_end":   fn.LineEnd,
		}

		// Thêm vào danh sách chờ Embed
		textsToEmbed = append(textsToEmbed, codeStr)
		chunks = append(chunks, port.CodeEmbedding{
			RepoName:     repoName,
			FilePath:     filePath,
			Language:     language,
			FunctionName: fn.Name,
			Content:      codeStr,
			CommitHash:   commitHash,
			Metadata:     meta,
		})
	}

	if len(textsToEmbed) == 0 {
		slog.Info("Không có hàm nào để embed trong file này", "file", filePath)
		return nil
	}

	// 3. Gọi Python AI để lấy Vector (Batch)
	slog.Info("Đang gọi AI lấy Vector...", "count", len(textsToEmbed))
	embeddings, err := s.aiClient.GetEmbeddings(ctx, textsToEmbed)
	if err != nil {
		return fmt.Errorf("lỗi lấy embedding: %w", err)
	}

	if len(embeddings) != len(chunks) {
		return fmt.Errorf("số lượng vector trả về (%d) không khớp với số lượng code (%d)", len(embeddings), len(chunks))
	}

	// 4. Gắn Vector vào Chunk
	for i := range chunks {
		chunks[i].Embedding = embeddings[i]
	}

	// 5. Lưu toàn bộ vào Postgres (F6)
	slog.Info("Đang lưu vào DB...", "count", len(chunks))
	err = s.vectorRepo.SaveEmbeddings(ctx, chunks)
	if err != nil {
		return fmt.Errorf("lỗi lưu DB: %w", err)
	}

	slog.Info("Xử lý file thành công!", "file", filePath)
	return nil
}

// Hàm helper để cắt code từ mảng các dòng
func extractCodeByLines(lines []string, start, end int32) string {
	// Kiểm tra biên (tránh panic index out of range)
	if start < 1 || int(end) > len(lines) || start > end {
		return ""
	}
	// Mảng trong Go bắt đầu từ 0, nên phải trừ 1
	return strings.Join(lines[start-1:end], "\n")
}
