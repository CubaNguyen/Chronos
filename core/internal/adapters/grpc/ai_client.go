package grpc

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/cuba3/chronos/core/internal/port"
	pb "github.com/cuba3/chronos/core/pkg/contracts/aiv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type aiClient struct {
	client pb.AIServiceClient
}

func NewAIClient(conn *grpc.ClientConn) port.AI {
	return &aiClient{
		client: pb.NewAIServiceClient(conn),
	}
}

func (c *aiClient) GetEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	req := &pb.EmbedRequest{Texts: texts}
	resp, err := c.client.Embed(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("lỗi gRPC (code: %s): %s", st.Code(), st.Message())
		}
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("AI embed thất bại: %s", resp.Error)
	}
	result := make([][]float32, 0, len(resp.Vectors))
	for _, v := range resp.Vectors {
		result = append(result, v.Values)
	}
	return result, nil
}

func (c *aiClient) Retrieve(ctx context.Context, query string, topK int, minScore float32) ([]port.ScoredChunk, error) {
	req := &pb.RetrieveRequest{Query: query, TopK: int32(topK), MinScore: minScore}
	resp, err := c.client.Retrieve(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("AI retrieve thất bại: %s", resp.Error)
	}
	return mapPbScoredChunks(resp.Results), nil
}

// 🚀 STREAMING FIXED HERE
func (c *aiClient) QueryStream(ctx context.Context, question string, topK int, minScore float32, contexts []port.Chunk) (<-chan *port.QueryResult, error) {

	// Map port.Chunk to pb.Chunk
	var pbContexts []*pb.Chunk
	for _, c := range contexts {
		pbContexts = append(pbContexts, &pb.Chunk{
			Id:         c.ID,
			Text:       c.Text,
			SourceType: c.SourceType,
			SourcePath: c.SourcePath,
			Metadata:   c.Metadata,
		})
	}

	req := &pb.QueryRequest{
		Question: question,
		TopK:     int32(topK),
		MinScore: minScore,
		Contexts: pbContexts, // 🚀 Pass mapped contexts
	}

	stream, err := c.client.Query(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start gRPC stream: %w", err)
	}

	resultChan := make(chan *port.QueryResult, 10)

	go func() {
		defer close(resultChan)
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break // Stream finished successfully
			}
			if err != nil {
				// 🚀 Nếu API Google lỗi 503, báo nhẹ nhàng thay vì crash
				slog.Error("Error receiving from gRPC stream", "error", err)
				// Gửi một câu trả lời lỗi về cho HTTP
				select {
				case resultChan <- &port.QueryResult{Answer: "\n[Lỗi: Server AI đang quá tải, vui lòng thử lại sau]"}:
				case <-ctx.Done():
				}
				break
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				slog.Error("Error receiving from gRPC stream", "error", err)
				break
			}

			result := &port.QueryResult{
				Answer:    resp.Answer,
				Sources:   mapPbScoredChunks(resp.Sources),
				Query:     resp.Query,
				NoContext: resp.NoContext,
			}

			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultChan, nil
}

func mapPbScoredChunks(pbChunks []*pb.ScoredChunk) []port.ScoredChunk {
	if pbChunks == nil {
		return nil
	}
	result := make([]port.ScoredChunk, 0, len(pbChunks))
	for _, sc := range pbChunks {
		result = append(result, port.ScoredChunk{
			Chunk: &port.Chunk{
				ID: sc.Chunk.Id, Text: sc.Chunk.Text, SourceType: sc.Chunk.SourceType,
				SourcePath: sc.Chunk.SourcePath, Metadata: sc.Chunk.Metadata,
			},
			Score: sc.Score,
		})
	}
	return result
}
