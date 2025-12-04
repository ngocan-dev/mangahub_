package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ngocan-dev/mangahub_/cli/client"
	"github.com/spf13/cobra"
)

var readChapterCmd = &cobra.Command{
	Use:   "read-chapter <manga-id> <chapter>",
	Short: "Đọc chapter và cập nhật tiến độ",
	Args:  cobra.ExactArgs(2),
	Run:   runReadChapter,
}

func init() {
	rootCmd.AddCommand(readChapterCmd)
}

func runReadChapter(cmd *cobra.Command, args []string) {
	mangaID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Println("❌ Manga ID không hợp lệ")
		os.Exit(1)
	}

	chapter, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Println("❌ Chapter không hợp lệ")
		os.Exit(1)
	}

	cfg, err := getStoredConfig()
	if err != nil || cfg.Token == "" {
		fmt.Println("❌ Chưa đăng nhập. Vui lòng chạy: mangahub login")
		os.Exit(1)
	}

	httpClient := client.NewHTTPClient(cfg.APIURL)
	httpClient.SetToken(cfg.Token)

	fmt.Printf("\n📖 Đang cập nhật tiến độ: Manga %d - Chapter %d...\n", mangaID, chapter)

	if err := httpClient.UpdateProgress(mangaID, chapter, nil); err != nil {
		fmt.Println("❌ Lỗi cập nhật tiến độ:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Đã cập nhật tiến độ thành công!")
	fmt.Println("\n💡 Tiến độ của bạn sẽ được đồng bộ đến các thiết bị khác qua TCP")
}
