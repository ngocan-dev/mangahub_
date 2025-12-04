package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ngocan-dev/mangahub_/cli/client"
	"github.com/spf13/cobra"
)

var showMangaCmd = &cobra.Command{
	Use:   "show-manga <id>",
	Short: "Xem chi tiết manga",
	Args:  cobra.ExactArgs(1),
	Run:   runShowManga,
}

func init() {
	rootCmd.AddCommand(showMangaCmd)
}

func runShowManga(cmd *cobra.Command, args []string) {
	mangaID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Println("❌ ID không hợp lệ")
		os.Exit(1)
	}

	cfg, err := getStoredConfig()
	if err != nil {
		fmt.Println("❌ Lỗi đọc config:", err)
		os.Exit(1)
	}

	httpClient := client.NewHTTPClient(cfg.APIURL)
	httpClient.SetToken(cfg.Token)

	detail, err := httpClient.GetMangaDetails(mangaID)
	if err != nil {
		fmt.Println("❌ Lỗi:", err)
		os.Exit(1)
	}

	fmt.Printf("\n📖 %s\n", detail.Title)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("ID: %d\n", detail.ID)
	fmt.Printf("Tác giả: %s\n", detail.Author)
	fmt.Printf("Thể loại: %s\n", detail.Genre)
	fmt.Printf("Trạng thái: %s\n", detail.Status)
	fmt.Printf("⭐ Đánh giá: %.1f/10\n", detail.RatingPoint)
	fmt.Printf("📚 Số chapter: %d\n\n", detail.ChapterCount)
	fmt.Printf("Mô tả:\n%s\n", detail.Description)

	if len(detail.Chapters) > 0 {
		fmt.Printf("\n📑 Chapters (hiển thị %d đầu):\n", len(detail.Chapters))
		for i, ch := range detail.Chapters {
			if i >= 5 {
				break
			}
			fmt.Printf("  • Chapter %d: %s\n", ch.ChapterNum, ch.Title)
		}
	}
}
