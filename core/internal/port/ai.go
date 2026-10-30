package port

import "context"

// === 1. DOMAIN MODELS (Chuyển hóa từ Protobuf) ===

// Chunk đại diện cho một đoạn code được tìm thấy
type Chunk struct {
	ID         string
	Text       string
	SourceType string // "function" | "method" | "class"
	SourcePath string
	Metadata   map[string]string
}

// ScoredChunk là Chunk kèm theo điểm số tương đồng (Cosine Similarity)
type ScoredChunk struct {
	Chunk *Chunk // Dùng con trỏ để tránh copy dữ liệu lớn
	Score float32
}

// QueryResult đại diện cho kết quả trả về của tính năng RAG (F7)
type QueryResult struct {
	Answer          string
	Sources         []ScoredChunk
	Query           string
	DurationSeconds float32
	NoContext       bool
}

// === 2. INTERFACE (Hợp đồng giao tiếp với Python) ===

type AI interface {
	// 1. Embed: Biến danh sách text thành danh sách vector (384 chiều)
	// (Dùng [][]float32 thay vì struct Vector cho chuẩn Go)
	GetEmbeddings(ctx context.Context, texts []string) ([][]float32, error)

	// 2. Retrieve: Tìm kiếm ngữ nghĩa thô (chỉ trả về code giống nhất, chưa qua LLM)
	Retrieve(ctx context.Context, query string, topK int, minScore float32) ([]ScoredChunk, error)

	// 3. Query (RAG): Hỏi đáp thông minh (Nhận contexts từ Postgres để LLM trả lời)
	Query(ctx context.Context, question string, topK int, minScore float32, contexts []Chunk) (*QueryResult, error)
}
