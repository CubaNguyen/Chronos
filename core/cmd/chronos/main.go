package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cuba3/chronos/core/internal/app"
	"github.com/cuba3/chronos/core/internal/config"
)

func main() {
	// 1. Đọc Config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Lỗi đọc config", "error", err)
		os.Exit(1)
	}

	// 2. Lắp ráp App
	application, err := app.NewApplication(cfg)
	if err != nil {
		slog.Error("Không thể khởi tạo app", "error", err)
		os.Exit(1)
	}
	// Đảm bảo tài nguyên được đóng khi tắt app
	defer application.Close()

	// 3. Bắt tín hiệu tắt máy (Ctrl+C)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 4. Chạy App
	if err := application.Run(ctx); err != nil {
		slog.Error("Lỗi khi chạy app", "error", err)
		os.Exit(1)
	}
}
