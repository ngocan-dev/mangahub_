package progress

import (
	"errors"

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
	Use:     "update",
	Short:   "Update reading progress",
	Long:    "Update reading progress for a specific manga and chapter.",
	Example: "mangahub progress update --manga-id one-piece --chapter 1095",
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		chapter, _ := cmd.Flags().GetInt("chapter")

		if mangaID == "" {
			return errors.New("--manga-id is required")
		}
		if chapter <= 0 {
			return errors.New("--chapter must be greater than 0")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}
		if cfg.Data.Token == "" {
			return errors.New("Please login first")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		payload, err := client.UpdateProgress(cmd.Context(), mangaID, chapter)
		if err != nil {
			return err
		}

		output.PrintJSON(cmd, payload)
		output.PrintSuccess(cmd, "Progress updated")
		return nil
	},
}

func init() {
	ProgressCmd.AddCommand(updateCmd)
	updateCmd.Flags().String("manga-id", "", "Manga identifier")
	updateCmd.Flags().Int("chapter", 0, "Chapter number")
	updateCmd.MarkFlagRequired("manga-id")
	updateCmd.MarkFlagRequired("chapter")
}
