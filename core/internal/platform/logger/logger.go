// internal/platform/logger/logger.go
package logger

import (
	"log/slog"
	"os"
)

// Init khởi tạo logger toàn cục
func Init(env string) *slog.Logger {
	var handler slog.Handler
	if env == "production" {
		// Log dạng JSON cho production (dễ parse bằng ELK/Datadog)
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		// Log dạng Text màu sắc cho dev
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger) // Set làm logger mặc định toàn cục
	return logger
}
