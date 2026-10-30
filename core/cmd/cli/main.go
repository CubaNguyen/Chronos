package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	// Tạo command gốc (root command)
	var rootCmd = &cobra.Command{
		Use:   "chronos",
		Short: "Chronos là công cụ phân tích và hỏi đáp mã nguồn",
		Long:  `Chronos giúp bạn index mã nguồn và đặt câu hỏi về code của mình.`,
	}

	// Thêm các lệnh con (ask, index) vào đây
	rootCmd.AddCommand(askCmd)
	rootCmd.AddCommand(indexCmd)

	// Chạy CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
