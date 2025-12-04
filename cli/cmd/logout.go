package cmd

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/config"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Đăng xuất khỏi MangaHub",
	Long:  `Xóa thông tin đăng nhập đã lưu`,
	Run:   runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) {
	// Xóa config
	if err := config.Clear(); err != nil {
		fmt.Printf("❌ Lỗi khi đăng xuất: %v\n", err)
		return
	}

	fmt.Println("✅ Đã đăng xuất thành công")
	fmt.Println("👋 Hẹn gặp lại!")
}
