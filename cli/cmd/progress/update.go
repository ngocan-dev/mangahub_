package progress

import (
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

// ProgressCmd groups reading progress commands.
var ProgressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Manage reading progress",
	Long:  "Update and synchronize reading progress across MangaHub services.",
}

// updateCmd updates progress for a manga.
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update reading progress",
	Long:  "Update reading progress for a specific manga and chapter.",
	Example: strings.Join([]string{
		"mangahub progress update --manga-id <id> --chapter <number>",
	}, "\n"),
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		chapter, _ := cmd.Flags().GetInt("chapter")

		if mangaID == "" {
			return fmt.Errorf("--manga-id is required")
		}
		if chapter <= 0 {
			return fmt.Errorf("--chapter must be greater than 0")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}
		if strings.TrimSpace(cfg.Data.Token) == "" {
			return fmt.Errorf("✗ You must be logged in to update progress.\nPlease login first.")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		resp, err := client.UpdateProgress(cmd.Context(), mangaID, chapter)
		if err != nil {
			return err
		}

		output.PrintJSON(cmd, resp)
		if config.Runtime().Quiet {
			return nil
		}

		cmd.Println("Updating reading progress...")
		cmd.Println("✓ Progress updated successfully.")
		cmd.Println("")
		cmd.Printf("Manga ID: %s\n", mangaID)
		if resp.UserProgress != nil {
			cmd.Printf("Current Chapter: %s\n", formatNumber(resp.UserProgress.CurrentChapter))
			cmd.Printf("Last Read: %s\n", resp.UserProgress.LastReadAt.UTC().Format("2006-01-02 15:04:05 MST"))
		}
		if resp.Broadcasted {
			cmd.Println("Notification: broadcasted to connected devices")
		} else {
			cmd.Println("Notification: not broadcasted")
		}
		if strings.TrimSpace(resp.Message) != "" {
			cmd.Printf("Message: %s\n", resp.Message)
		}
		cmd.Println("")
		return nil
	},
}

func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre == 0 {
		pre = 3
	}
	b.WriteString(s[:pre])
	for i := pre; i < len(s); i += 3 {
		b.WriteString(",")
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

func init() {
	ProgressCmd.AddCommand(updateCmd)
	updateCmd.Flags().String("manga-id", "", "Manga identifier")
	updateCmd.Flags().Int("chapter", 0, "Chapter number")
	updateCmd.MarkFlagRequired("manga-id")
	updateCmd.MarkFlagRequired("chapter")
}
