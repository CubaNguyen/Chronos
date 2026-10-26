package parser

import (
	"context"
	"fmt"
	"time"

	"github.com/cuba3/chronos/core/pkg/contracts/parserv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client parserv1.ParserServiceClient
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to parser: %w", err)
	}

	return &Client{
		conn:   conn,
		client: parserv1.NewParserServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) ParseFile(ctx context.Context, path, source string) (*parserv1.Ast, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.client.ParseFile(ctx, &parserv1.ParseFileRequest{
		Path:   path,
		Source: source,
	})
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("parser error: %s", resp.Error)
	}

	return resp.Ast, nil
}