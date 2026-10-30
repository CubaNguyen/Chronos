package grpc

import (
	"context"
	"fmt"

	"github.com/cuba3/chronos/core/internal/port"
	pb "github.com/cuba3/chronos/core/pkg/contracts/aiv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type aiClient struct {
	client pb.AIServiceClient
}

// NewAIClient khởi tạo Adapter và trả về Interface port.AI
func NewAIClient(conn *grpc.ClientConn) port.AI {
	return &aiClient{
		client: pb.NewAIServiceClient(conn),
	}
}

// =========================================================================
// 1. EMBED (GetEmbeddings)
// =========================================================================
func (c *aiClient) GetEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	req := &pb.EmbedRequest{
		Texts: texts,
	}

	resp, err := c.client.Embed(ctx, req)
	if err != nil {
		// Xử lý lỗi hệ thống (gRPC)
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("lỗi gRPC từ AI Embed (code: %s): %s", st.Code(), st.Message())
		}
		return nil, fmt.Errorf("lỗi không xác định khi gọi AI Embed: %w", err)
	}

	// Xử lý lỗi nghiệp vụ (Python trả về trong response)
	if resp.Error != "" {
		return nil, fmt.Errorf("AI embed thất bại: %s", resp.Error)
	}

	// MAPPING: []*pb.Vector -> [][]float32
	// Chúng ta bóc lớp vỏ Vector đi, chỉ lấy mảng float32
	result := make([][]float32, 0, len(resp.Vectors))
	for _, v := range resp.Vectors {
		result = append(result, v.Values)
	}

	return result, nil
}

// =========================================================================
// 2. RETRIEVE
// =========================================================================
func (c *aiClient) Retrieve(ctx context.Context, query string, topK int, minScore float32) ([]port.ScoredChunk, error) {
	req := &pb.RetrieveRequest{
		Query:    query,
		TopK:     int32(topK),
		MinScore: minScore,
	}

	resp, err := c.client.Retrieve(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, fmt.Errorf("lỗi gRPC từ AI Retrieve (code: %s): %s", st.Code(), st.Message())
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("AI retrieve thất bại: %s", resp.Error)
	}

	// MAPPING: []*pb.ScoredChunk -> []port.ScoredChunk
	return mapPbScoredChunks(resp.Results), nil
}

// =========================================================================
// 3. QUERY (RAG)
// =========================================================================
func (c *aiClient) Query(ctx context.Context, question string, topK int, minScore float32, contexts []port.Chunk) (*port.QueryResult, error) {
	pbContexts := make([]*pb.Chunk, 0, len(contexts))
	for _, item := range contexts {
		pbContexts = append(pbContexts, &pb.Chunk{
			Id:         item.ID,
			Text:       item.Text,
			SourceType: item.SourceType,
			SourcePath: item.SourcePath,
			Metadata:   item.Metadata,
		})
	}

	req := &pb.QueryRequest{
		Question: question,
		TopK:     int32(topK),
		MinScore: minScore,
		Contexts: pbContexts,
	}

	resp, err := c.client.Query(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		return nil, fmt.Errorf("lỗi gRPC từ AI Query (code: %s): %s", st.Code(), st.Message())
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("AI query thất bại: %s", resp.Error)
	}

	// MAPPING: pb.QueryResponse -> port.QueryResult
	result := &port.QueryResult{
		Answer:          resp.Answer,
		Sources:         mapPbScoredChunks(resp.Sources),
		Query:           resp.Query,
		DurationSeconds: resp.DurationSeconds,
		NoContext:       resp.NoContext,
	}

	return result, nil
}

// =========================================================================
// CÁC HÀM HELPER MAPPING (Tái sử dụng)
// =========================================================================

// mapPbScoredChunks map một mảng từ Protobuf sang Domain
func mapPbScoredChunks(pbChunks []*pb.ScoredChunk) []port.ScoredChunk {
	if pbChunks == nil {
		return nil
	}

	result := make([]port.ScoredChunk, 0, len(pbChunks))
	for _, sc := range pbChunks {
		result = append(result, port.ScoredChunk{
			Chunk: mapPbChunk(sc.Chunk),
			Score: sc.Score,
		})
	}
	return result
}

// mapPbChunk map một Chunk từ Protobuf sang Domain
func mapPbChunk(pbChunk *pb.Chunk) *port.Chunk {
	if pbChunk == nil {
		return nil
	}

	return &port.Chunk{
		ID:         pbChunk.Id,
		Text:       pbChunk.Text,
		SourceType: pbChunk.SourceType,
		SourcePath: pbChunk.SourcePath,
		Metadata:   pbChunk.Metadata,
	}
}
