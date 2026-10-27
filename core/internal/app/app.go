package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	dbAdapter "github.com/cuba3/chronos/core/internal/adapters/db"
	grpcAdapter "github.com/cuba3/chronos/core/internal/adapters/grpc"
	"github.com/cuba3/chronos/core/internal/config"
	"github.com/cuba3/chronos/core/internal/port"
)

type Application struct {
	Config       *config.Config
	DBPool       *pgxpool.Pool         // Giữ lại để đóng (Close)
	ParserConn   *grpc.ClientConn      // Giữ lại để đóng (Close)
	VectorRepo   port.VectorRepository // Repo DB
	ParserClient port.Parser           // Client Parser
}

// NewApplication là hàm "Lắp ráp" (Composition Root)
func NewApplication(cfg *config.Config) (*Application, error) {
	app := &Application{
		Config: cfg,
	}

	// --- 1. KHỞI TẠO HẠ TẦNG (Adapters) ---

	// Khởi tạo DB
	pool, err := dbAdapter.NewPostgresConn(cfg.DB.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("khởi tạo Postgres thất bại: %w", err)
	}
	app.DBPool = pool

	// Khởi tạo gRPC (Rust Parser)
	parserConn, err := grpcAdapter.NewParserConn(cfg.GRPC.RustParserAddr)
	if err != nil {
		app.DBPool.Close() // Nếu gRPC lỗi, đóng DB để tránh rò rỉ
		return nil, fmt.Errorf("khởi tạo gRPC Parser thất bại: %w", err)
	}
	app.ParserConn = parserConn

	app.VectorRepo = dbAdapter.NewPostgresRepo(pool)
	app.ParserClient = grpcAdapter.NewParserClient(parserConn)

	slog.Info("Đã lắp ráp xong các Adapters và Services!")
	return app, nil
}

func (a *Application) Run(ctx context.Context) error {
	slog.Info("Ứng dụng Chronos đang chạy...")

	<-ctx.Done()
	slog.Info("Nhận tín hiệu tắt máy, đang dọn dẹp...")

	return nil
}

// Close là hàm dọn dẹp tài nguyên (Rất quan trọng)
func (a *Application) Close() {
	if a.ParserConn != nil {
		a.ParserConn.Close()
		slog.Info("Đã đóng kết nối gRPC")
	}
	if a.DBPool != nil {
		a.DBPool.Close()
		slog.Info("Đã đóng kết nối Database")
	}
}
