// internal/adapters/db/connection.go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresConn(connString string) (*pgxpool.Pool, error) {
	// Cấu hình Pool (Rất quan trọng để chịu tải)
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("lỗi parse config DB: %w", err)
	}

	// MaxConns: Số kết nối tối đa (tùy thuộc cấu hình server Postgres)
	config.MaxConns = 10
	// MaxConnLifetime: Thời gian sống tối đa của 1 connection (tránh bị firewall kill)
	config.MaxConnLifetime = time.Hour

	// Tạo context timeout để không bị treo nếu DB sập
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối tới Postgres: %w", err)
	}

	// Ping thử xem DB có thực sự phản hồi không
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("DB không phản hồi: %w", err)
	}

	return pool, nil
}
