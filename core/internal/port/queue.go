package port

import "context"

type Job struct {
	RepoName   string `json:"repo_name"`
	FilePath   string `json:"file_path"`
	Language   string `json:"language"`
	SourceCode string `json:"source_code"`
}

// Queue định nghĩa hợp đồng gửi/nhận message
type Queue interface {
	PublishJob(ctx context.Context, job Job) error
	Close()
}
