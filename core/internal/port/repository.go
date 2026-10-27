package port

import "context"

// 1. Cấu trúc dữ liệu (Domain Model)
type CodeEmbedding struct {
	ID           string
	RepoName     string
	FilePath     string
	Language     string
	FunctionName string
	Content      string
	Embedding    []float32
	CommitHash   string
	Metadata     map[string]interface{} // Dùng map để chứa JSONB
}

// 2. Interface (Hợp đồng)
type VectorRepository interface {
	SaveEmbeddings(ctx context.Context, items []CodeEmbedding) error
	SearchSimilar(ctx context.Context, queryEmbedding []float32, limit int) ([]CodeEmbedding, error)
}
