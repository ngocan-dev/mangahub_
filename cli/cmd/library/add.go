package library

import (
	"errors"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

// LibraryCmd manages library entries.
var LibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Manage your MangaHub library",
	Long:  "Add, remove, and update manga entries in your personal library.",
}

// addCmd adds a manga entry to the user's library.
var addCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add a manga to the library",
	Long:    "Add a manga to your library by ID with an optional status.",
	Example: "mangahub library add --manga-id one-piece --status reading",
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		status, _ := cmd.Flags().GetString("status")

		if mangaID == "" {
			return errors.New("--manga-id is required")
		}
		if status == "" {
			status = "reading"
		}
		switch status {
		case "reading", "completed", "planned":
		default:
			return errors.New("--status must be one of: reading, completed, planned")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}
		if cfg.Data.Token == "" {
			return errors.New("Please login first")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		payload, err := client.AddToLibrary(cmd.Context(), mangaID, status)
		if err != nil {
			return err
		}

		output.PrintJSON(cmd, payload)
		output.PrintSuccess(cmd, "Manga added to library")
		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(addCmd)
	addCmd.Flags().String("manga-id", "", "Manga identifier")
	addCmd.Flags().String("status", "reading", "Reading status (reading|completed|planned)")
	addCmd.MarkFlagRequired("manga-id")
}
