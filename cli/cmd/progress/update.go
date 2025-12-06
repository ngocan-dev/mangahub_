package progress

import (
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	isync "github.com/ngocan-dev/mangahub_/cli/internal/sync"
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
		"mangahub progress update --manga-id <id> --chapter <number> --volume <number>",
		"mangahub progress update --manga-id <id> --chapter <number> --notes \"message\"",
	}, "\n"),
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		chapter, _ := cmd.Flags().GetInt("chapter")
		volume, _ := cmd.Flags().GetInt("volume")
		notes, _ := cmd.Flags().GetString("notes")
		force, _ := cmd.Flags().GetBool("force")

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
		req := api.ProgressUpdateRequest{MangaID: mangaID, Chapter: chapter, Volume: volume, Notes: notes, Force: force}

		resp, err := client.UpdateProgressDetail(cmd.Context(), req)
		if err != nil {
			return err
		}

		output.PrintJSON(cmd, resp)
		if config.Runtime().Quiet {
			output.PrintSuccess(cmd, "✓ Progress updated successfully!")
			return nil
		}

		delta := resp.CurrentChapter - resp.PreviousChapter
		deltaStr := fmt.Sprintf("(%+d)", delta)

		cmd.Println("Updating reading progress...")
		cmd.Println("✓ Progress updated successfully!")
		cmd.Println("")
		cmd.Printf("Manga: %s\n", resp.MangaName)
		cmd.Printf("Previous: Chapter %s\n", formatNumber(resp.PreviousChapter))
		cmd.Printf("Current: Chapter %s %s\n", formatNumber(resp.CurrentChapter), deltaStr)
		cmd.Printf("Updated: %s\n", resp.UpdatedAt.UTC().Format("2006-01-02 15:04:05 MST"))
		cmd.Println("")

		cmd.Println("Sync Status:")
		cmd.Printf("Local database: %s %s\n", icon(resp.Sync.Local.Updated), resp.Sync.Local.Note)

		tcpResult := isync.Broadcast(resp.Sync.TCP.Devices)
		tcpIcon := "✗"
		tcpMsg := "Failed"
		if tcpResult.Error == nil {
			tcpIcon = "✓"
			tcpMsg = fmt.Sprintf("Broadcasting to %d connected devices", tcpResult.Devices)
		}
		cmd.Printf("TCP sync server: %s %s\n", tcpIcon, tcpMsg)

		cloudIcon := icon(resp.Sync.Cloud.Success)
		cloudMsg := resp.Sync.Cloud.Message
		cmd.Printf("Cloud backup: %s %s\n", cloudIcon, cloudMsg)
		cmd.Println("")

		cmd.Println("Statistics:")
		cmd.Printf("Total chapters read: %s\n", formatNumber(resp.Statistics.TotalRead))
		cmd.Printf("Reading streak: %d days\n", resp.Statistics.ReadingStreak)
		cmd.Printf("Estimated completion: %s\n", resp.Statistics.EstimatedCompletion)
		cmd.Println("")

		cmd.Println("Next actions:")
		cmd.Printf("Continue reading: Chapter %s available\n", formatNumber(resp.CurrentChapter+1))
		cmd.Printf("Rate this chapter: mangahub library update --manga-id %s --rating 9\n", resp.MangaID)
		return nil
	},
}

func icon(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
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
	updateCmd.Flags().Int("volume", 0, "Volume number")
	updateCmd.Flags().String("notes", "", "Optional notes")
	updateCmd.Flags().Bool("force", false, "Allows backward chapter updates")
	updateCmd.MarkFlagRequired("manga-id")
	updateCmd.MarkFlagRequired("chapter")
}
