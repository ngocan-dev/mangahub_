package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ngocan-dev/mangahub_/cli/client"
	"github.com/spf13/cobra"
)

var syncProgressCmd = &cobra.Command{
	Use:   "sync-progress",
	Short: "Đồng bộ tiến độ đọc realtime qua TCP",
	Long:  `Kết nối với TCP server để nhận cập nhật tiến độ đọc từ các thiết bị khác`,
	Run:   runSyncProgress,
}

var (
	tcpServerAddr string
)

func init() {
	rootCmd.AddCommand(syncProgressCmd)

	syncProgressCmd.Flags().StringVar(&tcpServerAddr, "tcp-server", "localhost:9000", "Địa chỉ TCP server")
}

func runSyncProgress(cmd *cobra.Command, args []string) {
	// Lấy token từ config
	token := getStoredToken()
	if token == "" {
		fmt.Println("❌ Chưa đăng nhập. Vui lòng chạy: mangahub login")
		os.Exit(1)
	}

	// Tạo TCP client
	tcpClient := client.NewTCPClient(tcpServerAddr, token)

	// Set callback cho progress updates
	tcpClient.SetProgressCallback(func(progress client.ProgressUpdate) {
		fmt.Printf("\n📖 Cập nhật tiến độ đọc!\n")
		fmt.Printf("   User ID: %d\n", progress.UserID)
		fmt.Printf("   Manga ID: %d\n", progress.NovelID)
		fmt.Printf("   Chapter: %d\n", progress.Chapter)
		if progress.ChapterID != nil {
			fmt.Printf("   Chapter ID: %d\n", *progress.ChapterID)
		}
		fmt.Printf("   Thời gian: %s\n\n", progress.Timestamp)
	})

	// Kết nối
	if err := tcpClient.Connect(); err != nil {
		fmt.Printf("❌ Lỗi kết nối TCP: %v\n", err)
		os.Exit(1)
	}
	defer tcpClient.Close()

	fmt.Println("\n✓ Đang đồng bộ tiến độ... (Nhấn Ctrl+C để thoát)")
	fmt.Println("💡 Khi bạn đọc manga trên thiết bị khác, tiến độ sẽ hiện ở đây")

	// Đợi signal để thoát
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n👋 Đã ngắt kết nối")
}
