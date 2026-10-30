package message

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cuba3/chronos/core/internal/port"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	StreamName    = "PIPELINE_JOBS"
	SubjectName   = "pipeline.jobs"
	ConsumerGroup = "pipeline-workers"
)

type NatsQueue struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func (q *NatsQueue) JS() jetstream.JetStream {
	return q.js
}

// NewNatsQueue khởi tạo kết nối và tạo Stream nếu chưa có
func NewNatsQueue(url string) (*NatsQueue, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối NATS: %w", err)
	}

	// Khởi tạo JetStream context
	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("không thể khởi tạo JetStream: %w", err)
	}

	ctx := context.Background()
	// Tạo Stream (giống tạo Topic bên Kafka) nếu chưa tồn tại
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     StreamName,
		Subjects: []string{SubjectName},
	})
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("không thể tạo Stream: %w", err)
	}

	return &NatsQueue{nc: nc, js: js}, nil
}

// PublishJob gửi một công việc vào NATS
func (q *NatsQueue) PublishJob(ctx context.Context, job port.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("lỗi marshal job: %w", err)
	}

	// Gửi message vào Subject
	_, err = q.js.Publish(ctx, SubjectName, data)
	if err != nil {
		return fmt.Errorf("lỗi publish message: %w", err)
	}
	return nil
}

func (q *NatsQueue) Close() {
	q.nc.Close()
}
