package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/cuba3/chronos/core/internal/app"
	"github.com/cuba3/chronos/core/internal/config"
	"github.com/spf13/cobra"
)

var (
	indexRepo string
	indexDir  string
	useGit    bool
)

var IndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Quét và đánh chỉ mục thư mục mã nguồn vào NATS",
	Run: func(cmd *cobra.Command, args []string) {
		if indexDir == "" {
			fmt.Println("Vui lòng nhập đường dẫn thư mục bằng cờ --dir hoặc -d")
			return
		}

		cfg, _ := config.Load()

		// Index chỉ cần NATS Queue để bắn job, không khởi tạo HTTP Server hay gRPC thừa
		indexerApp, err := app.NewIndexerApp(cfg)
		if err != nil {
			log.Fatalf("Lỗi khởi tạo app: %v", err)
		}
		defer indexerApp.Close()

		ctx := context.Background()
		if indexRepo == "" {
			indexRepo = "chronos-self"
		}

		fmt.Printf("Bắt đầu quét thư mục %s cho repo %s...\n", indexDir, indexRepo)
		err = indexerApp.GitIngestService.IngestDirectory(ctx, indexRepo, indexDir)
		if err != nil {
			log.Fatalf("Lỗi index thư mục: %v", err)
		}

		fmt.Println("Đã gửi toàn bộ files vào hàng đợi NATS thành công!")
	},
}

func init() {
	IndexCmd.Flags().StringVarP(&indexRepo, "repo", "r", "chronos-self", "Tên dự án")
	IndexCmd.Flags().StringVarP(&indexDir, "dir", "d", "", "Đường dẫn thư mục cần quét")
	IndexCmd.Flags().BoolVarP(&useGit, "git", "g", false, "Đọc từ lịch sử Git thay vì quét thư mục")
	IndexCmd.MarkFlagRequired("dir")
}
