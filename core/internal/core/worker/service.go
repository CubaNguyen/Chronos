package worker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"

	"github.com/cuba3/chronos/core/internal/core/ingest"
	"github.com/cuba3/chronos/core/internal/port"
	"github.com/nats-io/nats.go/jetstream"
)

type Worker struct {
	js            jetstream.JetStream
	ingestService *ingest.Service
}

func NewWorker(js jetstream.JetStream, ingestSvc *ingest.Service) *Worker {
	return &Worker{
		js:            js,
		ingestService: ingestSvc,
	}
}

// Start khởi chạy vòng lặp lắng nghe message
func (w *Worker) Start(ctx context.Context) error {
	// 1. Tạo Durable Consumer (Giống Consumer Group bên Kafka)
	cons, err := w.js.CreateOrUpdateConsumer(ctx, "PIPELINE_JOBS", jetstream.ConsumerConfig{
		Durable:   "pipeline-workers",
		AckPolicy: jetstream.AckExplicitPolicy, // Phải Ack thủ công
	})
	if err != nil {
		return err
	}

	slog.Info("Worker đang lắng nghe job từ NATS...")

	// 2. Lắng nghe message liên tục
	_, err = cons.Consume(func(msg jetstream.Msg) {
		var job port.Job
		if err := json.Unmarshal(msg.Data(), &job); err != nil {
			slog.Error("Lỗi parse JSON job", "error", err)
			msg.Nak() // Báo lỗi, NATS sẽ gửi lại sau
			return
		}

		slog.Info("Nhận được job mới", "file", job.FilePath)

		// Tạo hash từ nội dung source code để làm commit/content hash độc nhất
		hash := sha256.Sum256([]byte(job.SourceCode))
		contentHash := hex.EncodeToString(hash[:])

		// 3. Gọi Core Logic xử lý (truyền contentHash)
		err := w.ingestService.ProcessFile(ctx, job.RepoName, job.FilePath, job.Language, job.SourceCode, contentHash)

		if err != nil {
			slog.Error("Xử lý job thất bại", "error", err)
			msg.Nak() // Xử lý lỗi -> Trả lại cho NATS
			return
		}

		// 4. Xử lý thành công -> Báo cho NATS xóa message
		msg.Ack()
		slog.Info("Xử lý job thành công", "file", job.FilePath)
	})

	return err
}
