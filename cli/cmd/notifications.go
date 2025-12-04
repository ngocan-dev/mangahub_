package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ngocan-dev/mangahub_/cli/client"
	"github.com/spf13/cobra"
)

var notificationsCmd = &cobra.Command{
	Use:   "notifications",
	Short: "Nhận thông báo chapter mới qua UDP",
	Long:  `Kết nối với UDP server để nhận thông báo realtime khi có chapter mới phát hành`,
	Run:   runNotifications,
}

var (
	udpServerAddr string
	novelIDs      []int64
	allNovels     bool
)

func init() {
	rootCmd.AddCommand(notificationsCmd)

	notificationsCmd.Flags().StringVar(&udpServerAddr, "udp-server", "localhost:9091", "Địa chỉ UDP server")
	notificationsCmd.Flags().Int64SliceVar(&novelIDs, "novels", []int64{}, "Danh sách ID manga muốn nhận thông báo (để trống = tất cả)")
	notificationsCmd.Flags().BoolVar(&allNovels, "all", true, "Nhận thông báo từ tất cả manga")
}

func runNotifications(cmd *cobra.Command, args []string) {
	// Lấy token và user ID từ config
	token := getStoredToken()
	if token == "" {
		fmt.Println("❌ Chưa đăng nhập. Vui lòng chạy: mangahub login")
		os.Exit(1)
	}

	userID := getStoredUserID()
	if userID == 0 {
		fmt.Println("❌ Không tìm thấy user ID")
		os.Exit(1)
	}

	// Tạo UDP client
	udpClient := client.NewUDPClient(udpServerAddr, token, userID)

	// Cấu hình subscription
	if len(novelIDs) > 0 {
		udpClient.SubscribeToNovels(novelIDs)
		fmt.Printf("📡 Đăng ký nhận thông báo cho %d manga\n", len(novelIDs))
	} else {
		udpClient.SubscribeToAll()
		fmt.Println("📡 Đăng ký nhận thông báo cho TẤT CẢ manga")
	}

	// Set callback
	udpClient.SetNotificationCallback(func(notif client.ChapterNotification) {
		fmt.Printf("\n🔔 Thông báo mới!\n")
		fmt.Printf("   Manga: %s (ID: %d)\n", notif.NovelName, notif.NovelID)
		fmt.Printf("   Chapter: %d\n", notif.Chapter)
		fmt.Printf("   Thời gian: %s\n\n", notif.Timestamp)
	})

	// Kết nối
	if err := udpClient.Connect(); err != nil {
		fmt.Printf("❌ Lỗi kết nối UDP: %v\n", err)
		os.Exit(1)
	}
	defer udpClient.Close()

	fmt.Println("\n✓ Đang lắng nghe thông báo... (Nhấn Ctrl+C để thoát)")

	// Đợi signal để thoát
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n👋 Đã ngắt kết nối")
}
