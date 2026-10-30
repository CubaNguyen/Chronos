package grpc

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// NewParserConn khởi tạo kết nối tới service Rust (Parser)
func NewParserConn(addr string) (*grpc.ClientConn, error) {
	return createGRPCConn(addr)
}

// NewAIConn khởi tạo kết nối tới service Python (AI)
func NewAIConn(addr string) (*grpc.ClientConn, error) {
	return createGRPCConn(addr)
}

func createGRPCConn(addr string) (*grpc.ClientConn, error) {
	kaParams := keepalive.ClientParameters{
		Time:                1 * time.Minute,
		Timeout:             5 * time.Second,
		PermitWithoutStream: false,
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(kaParams),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 5 * time.Second,
		}),
	)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
