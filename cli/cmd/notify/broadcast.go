package notify

import (
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var broadcastCmd = &cobra.Command{
	Use:     "broadcast",
	Short:   "Broadcast a chapter release notification",
	Long:    "Trigger a chapter release notification broadcast to registered UDP clients.",
	Example: "mangahub notify broadcast --manga-id 42 --chapter 1096",
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetInt64("manga-id")
		chapter, _ := cmd.Flags().GetInt("chapter")
		chapterID, _ := cmd.Flags().GetInt64("chapter-id")

		if mangaID <= 0 {
			return fmt.Errorf("--manga-id must be greater than 0")
		}
		if chapter <= 0 {
			return fmt.Errorf("--chapter must be greater than 0")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}
		if strings.TrimSpace(cfg.Data.Token) == "" {
			return fmt.Errorf("✗ You must be logged in to broadcast notifications.\nPlease login first.")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		req := api.ChapterReleaseNotificationRequest{
			NovelID:   mangaID,
			Chapter:   chapter,
			ChapterID: chapterID,
		}
		resp, err := client.NotifyChapterRelease(cmd.Context(), req)
		if err != nil {
			return err
		}

		output.PrintJSON(cmd, resp)
		if config.Runtime().Quiet {
			cmd.Println("ok")
			return nil
		}

		cmd.Println("✓ Chapter release notification broadcasted.")
		cmd.Printf("Manga: %s (ID %d)\n", resp.NovelName, resp.NovelID)
		cmd.Printf("Chapter: %d\n", resp.Chapter)
		if resp.ChapterID > 0 {
			cmd.Printf("Chapter ID: %d\n", resp.ChapterID)
		}
		return nil
	},
}

func init() {
	NotifyCmd.AddCommand(broadcastCmd)
	broadcastCmd.Flags().Int64("manga-id", 0, "Manga identifier")
	broadcastCmd.Flags().Int("chapter", 0, "Chapter number")
	broadcastCmd.Flags().Int64("chapter-id", 0, "Optional chapter identifier")
	broadcastCmd.MarkFlagRequired("manga-id")
	broadcastCmd.MarkFlagRequired("chapter")
	output.AddFlag(broadcastCmd)
}
