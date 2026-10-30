package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	dbAdapter "github.com/cuba3/chronos/core/internal/adapters/db"
	grpcAdapter "github.com/cuba3/chronos/core/internal/adapters/grpc"
	httpAdapter "github.com/cuba3/chronos/core/internal/adapters/http"
	messageAdapter "github.com/cuba3/chronos/core/internal/adapters/message"
	"github.com/cuba3/chronos/core/internal/config"
	"github.com/cuba3/chronos/core/internal/core/ingest"
	"github.com/cuba3/chronos/core/internal/core/pipeline"
	"github.com/cuba3/chronos/core/internal/core/query"
	"github.com/cuba3/chronos/core/internal/core/search"
	"github.com/cuba3/chronos/core/internal/core/worker"
	"github.com/cuba3/chronos/core/internal/port"
)

type Application struct {
	Config       *config.Config
	DBPool       *pgxpool.Pool
	ParserConn   *grpc.ClientConn
	VectorRepo   port.VectorRepository
	ParserClient port.Parser
	Queue        port.Queue
	Worker       *worker.Worker

	PipelineService  *pipeline.Service
	QueryService     *query.Service
	SearchService    *search.Service
	IngestService    *ingest.Service
	GitIngestService *ingest.GitIngestService

	AIConn     *grpc.ClientConn
	AIClient   port.AI
	HTTPServer *http.Server
}

func NewApplication(cfg *config.Config) (*Application, error) {
	app := &Application{Config: cfg}

	// --- 1. HẠ TẦNG (Postgres, gRPC Parser, gRPC AI, NATS) ---
	pool, err := dbAdapter.NewPostgresConn(cfg.DB.PostgresURL)
	if err != nil {
		return nil, err
	}
	app.DBPool = pool

	parserConn, err := grpcAdapter.NewParserConn(cfg.GRPC.RustParserAddr)
	if err != nil {
		app.DBPool.Close()
		return nil, err
	}
	app.ParserConn = parserConn

	aiConn, err := grpcAdapter.NewAIConn(cfg.GRPC.PythonAIAddr)
	if err != nil {
		app.DBPool.Close()
		app.ParserConn.Close()
		return nil, fmt.Errorf("khởi tạo gRPC AI thất bại: %w", err)
	}
	app.AIConn = aiConn
	app.AIClient = grpcAdapter.NewAIClient(aiConn)

	// Khởi tạo NATS Queue
	natsQueue, err := messageAdapter.NewNatsQueue(cfg.NATS.URL)
	if err != nil {
		app.DBPool.Close()
		app.ParserConn.Close()
		app.AIConn.Close()
		return nil, fmt.Errorf("khởi tạo NATS thất bại: %w", err)
	}
	app.Queue = natsQueue

	// --- 2. ADAPTERS ---
	app.VectorRepo = dbAdapter.NewPostgresRepo(pool)
	app.ParserClient = grpcAdapter.NewParserClient(parserConn)

	// --- 3. CORE SERVICES ---
	app.PipelineService = pipeline.NewService(app.ParserClient)
	app.QueryService = query.NewService(app.AIClient, app.VectorRepo)
	app.SearchService = search.NewService(app.AIClient, app.VectorRepo)
	app.IngestService = ingest.NewService(app.ParserClient, app.AIClient, app.VectorRepo)
	app.GitIngestService = ingest.NewGitIngestService(app.Queue)

	// Worker lắng nghe job từ NATS và điều phối qua IngestService
	app.Worker = worker.NewWorker(natsQueue.JS(), app.IngestService)

	// --- 4. HTTP API ---
	pipelineHandler := httpAdapter.NewHandler(app.PipelineService)
	queryHandler := httpAdapter.NewQueryHandler(app.QueryService)
	searchHandler := httpAdapter.NewSearchHandler(app.SearchService)
	ingestHandler := httpAdapter.NewIngestHandler(app.IngestService, app.Queue)
	gitHandler := httpAdapter.NewGitHandler(app.GitIngestService)

	router := httpAdapter.SetupRouter(pipelineHandler, queryHandler, searchHandler, ingestHandler, gitHandler)

	app.HTTPServer = &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: router,
	}

	slog.Info("Đã lắp ráp xong các Adapters, Services, NATS Queue và Worker!")
	return app, nil
}

func (a *Application) Run(ctx context.Context) error {
	// 1. Chạy Worker ngầm trong 1 goroutine
	go func() {
		if err := a.Worker.Start(ctx); err != nil {
			slog.Error("Worker NATS bị lỗi", "error", err)
		}
	}()

	// 2. Chạy HTTP Server trong 1 Goroutine ngầm
	go func() {
		slog.Info("HTTP Server đang lắng nghe tại", "port", a.Config.HTTP.Port)
		if err := a.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Lỗi HTTP Server", "error", err)
		}
	}()

	// 3. Chờ tín hiệu tắt máy (Ctrl+C)
	<-ctx.Done()
	slog.Info("Nhận tín hiệu tắt máy, đang dọn dẹp...")

	// 4. Tắt HTTP Server an toàn (Graceful Shutdown 5s)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.HTTPServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("Lỗi khi tắt HTTP Server", "error", err)
	}

	return nil
}

// Close là hàm dọn dẹp tài nguyên
func (a *Application) Close() {
	if a.Queue != nil {
		a.Queue.Close()
		slog.Info("Đã đóng kết nối NATS")
	}
	if a.ParserConn != nil {
		a.ParserConn.Close()
		slog.Info("Đã đóng kết nối gRPC Parser")
	}
	if a.DBPool != nil {
		a.DBPool.Close()
		slog.Info("Đã đóng kết nối Database")
	}
	if a.AIConn != nil {
		a.AIConn.Close()
		slog.Info("Đã đóng kết nối gRPC AI")
	}
}
