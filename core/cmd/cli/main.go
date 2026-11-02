package main

import (
	"fmt"
	"os"

	"github.com/cuba3/chronos/core/cmd/cli/commands" // Import thư mục commands
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "chronos",
		Short: "Chronos là công cụ phân tích và hỏi đáp mã nguồn",
		Long:  `Chronos giúp bạn index mã nguồn và đặt câu hỏi về code của mình.`,
	}

	// Lắp ráp các lệnh vào
	rootCmd.AddCommand(commands.AskCmd)
	rootCmd.AddCommand(commands.IndexCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
