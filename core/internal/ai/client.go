package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/cuba3/chronos/core/pkg/contracts/aiv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client aiv1.AIServiceClient
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AI: %w", err)
	}

	return &Client{
		conn:   conn,
		client: aiv1.NewAIServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Query(ctx context.Context, question string, topK int32) (*aiv1.QueryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return c.client.Query(ctx, &aiv1.QueryRequest{
		Question: question,
		TopK:     topK,
	})
}