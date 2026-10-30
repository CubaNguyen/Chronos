package main

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
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Quét và đánh chỉ mục thư mục mã nguồn vào NATS",
	Run: func(cmd *cobra.Command, args []string) {
		if indexDir == "" {
			fmt.Println("Vui lòng nhập đường dẫn thư mục bằng cờ --dir hoặc -d")
			return
		}

		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Lỗi load config: %v", err)
		}

		application, err := app.NewApplication(cfg)
		if err != nil {
			log.Fatalf("Lỗi khởi tạo app: %v", err)
		}
		defer application.Close()

		ctx := context.Background()
		if indexRepo == "" {
			indexRepo = "chronos-self"
		}

		fmt.Printf("Bắt đầu quét thư mục %s cho repo %s...\n", indexDir, indexRepo)
		err = application.GitIngestService.IngestDirectory(ctx, indexRepo, indexDir)
		if err != nil {
			log.Fatalf("Lỗi index thư mục: %v", err)
		}

		fmt.Println("Đã gửi toàn bộ files vào hàng đợi NATS thành công!")
	},
}

func init() {
	indexCmd.Flags().StringVarP(&indexRepo, "repo", "r", "chronos-self", "Tên dự án")
	indexCmd.Flags().StringVarP(&indexDir, "dir", "d", "", "Đường dẫn thư mục cần quét")
	indexCmd.MarkFlagRequired("dir")
}
