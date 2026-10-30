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
	askRepo     string
	askQuestion string
)

var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Đặt câu hỏi về mã nguồn",
	Run: func(cmd *cobra.Command, args []string) {
		if askQuestion == "" {
			fmt.Println("Vui lòng nhập câu hỏi bằng cờ --question hoặc -q")
			return
		}

		// 1. Khởi tạo Config
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("Lỗi load config: %v", err)
		}

		// 2. Khởi tạo App (chỉ cần các service liên quan, không cần HTTP Server)
		// Tạm thời chúng ta sẽ dùng lại hàm NewApplication và lấy ra QueryService
		application, err := app.NewApplication(cfg)
		if err != nil {
			log.Fatalf("Lỗi khởi tạo app: %v", err)
		}
		defer application.Close()

		// 3. Gọi thẳng vào tầng Core Logic
		// (Giả sử application có lưu QueryService)
		ctx := context.Background()
		result, err := application.QueryService.Ask(ctx, askRepo, askQuestion)
		if err != nil {
			log.Fatalf("Lỗi truy vấn: %v", err)
		}

		// 4. In kết quả ra màn hình đẹp mắt
		fmt.Println("\n--- Câu trả lời ---")
		fmt.Println(result.Answer)
		fmt.Println("\n--- Nguồn tham khảo ---")
		for i, src := range result.Sources {
			fmt.Printf("[%d] %s (%s)\n", i+1, src.Chunk.SourcePath, src.Chunk.SourceType)
		}
	},
}

func init() {
	// Khai báo các cờ (flags)
	askCmd.Flags().StringVarP(&askRepo, "repo", "r", "", "Tên dự án cần hỏi")
	askCmd.Flags().StringVarP(&askQuestion, "question", "q", "", "Câu hỏi của bạn")

	// Bắt buộc phải có cờ question
	askCmd.MarkFlagRequired("question")
}
