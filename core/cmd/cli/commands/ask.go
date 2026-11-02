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
	askRepo     string
	askQuestion string
)

// 🚨 Viết hoa chữ Ask để main.go gọi được
var AskCmd = &cobra.Command{
	Use:   "ask",
	Short: "Đặt câu hỏi về mã nguồn",
	Run: func(cmd *cobra.Command, args []string) {
		if askQuestion == "" {
			fmt.Println("Vui lòng nhập câu hỏi bằng cờ --question hoặc -q")
			return
		}

		cfg, _ := config.Load()

		// 🚀 SỬA LỖI: Chỉ dùng NewCoreApp để không khởi tạo HTTP/Gin
		coreApp, err := app.NewCoreApp(cfg)
		if err != nil {
			log.Fatalf("Lỗi khởi tạo app: %v", err)
		}
		defer coreApp.Close()

		ctx := context.Background()
		result, err := coreApp.QueryService.Ask(ctx, askRepo, askQuestion)
		if err != nil {
			log.Fatalf("Lỗi truy vấn: %v", err)
		}

		fmt.Println("\n--- Câu trả lời ---")
		fmt.Println(result.Answer)
		fmt.Println("\n--- Nguồn tham khảo ---")
		for i, src := range result.Sources {
			fmt.Printf("[%d] %s (%s)\n", i+1, src.Chunk.SourcePath, src.Chunk.SourceType)
		}
	},
}

func init() {
	AskCmd.Flags().StringVarP(&askRepo, "repo", "r", "", "Tên dự án cần hỏi")
	AskCmd.Flags().StringVarP(&askQuestion, "question", "q", "", "Câu hỏi của bạn")
	AskCmd.MarkFlagRequired("question")
}
