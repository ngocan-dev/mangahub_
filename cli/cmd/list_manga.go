package cmd

import (
	"fmt"
	"os"

	"github.com/ngocan-dev/mangahub_/cli/client"
	"github.com/spf13/cobra"
)

var listMangaCmd = &cobra.Command{
	Use:   "list-manga",
	Short: "Liệt kê danh sách manga phổ biến",
	Run:   runListManga,
}

var (
	limit int
)

func init() {
	rootCmd.AddCommand(listMangaCmd)
	listMangaCmd.Flags().IntVar(&limit, "limit", 10, "Số lượng manga")
}

func runListManga(cmd *cobra.Command, args []string) {
	cfg, err := getStoredConfig()
	if err != nil {
		fmt.Println("❌ Lỗi đọc config:", err)
		os.Exit(1)
	}

	httpClient := client.NewHTTPClient(cfg.APIURL)
	httpClient.SetToken(cfg.Token)

	resp, err := httpClient.GetPopularManga(limit)
	if err != nil {
		fmt.Println("❌ Lỗi:", err)
		os.Exit(1)
	}

	fmt.Printf("\n📚 Top %d Manga phổ biến:\n\n", limit)
	for i, manga := range resp.Results {
		fmt.Printf("%d. [%d] %s\n", i+1, manga.ID, manga.Title)
		fmt.Printf("   Tác giả: %s | Thể loại: %s\n", manga.Author, manga.Genre)
		fmt.Printf("   ⭐ %.1f | Trạng thái: %s\n\n", manga.RatingPoint, manga.Status)
	}
}
