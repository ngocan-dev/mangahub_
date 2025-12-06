package library

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update a library entry",
	Long:    "Update progress or metadata for a manga in your library.",
	Example: "mangahub library update --manga-id one-piece --status completed --rating 10",
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		status, _ := cmd.Flags().GetString("status")
		ratingVal, _ := cmd.Flags().GetInt("rating")

		if strings.TrimSpace(mangaID) == "" {
			return errors.New("--manga-id is required")
		}
		status = strings.ToLower(strings.TrimSpace(status))
		if status == "" {
			return errors.New("--status is required")
		}
		if err := validateStatus(status); err != nil {
			cmd.Println("✗ Invalid status. Must be one of:")
			cmd.Println("reading, completed, plan-to-read, on-hold, dropped")
			return nil
		}

		var rating *int
		if cmd.Flags().Changed("rating") {
			if ratingVal < 1 || ratingVal > 10 {
				cmd.Println("✗ Invalid rating. Must be between 1 and 10.")
				return nil
			}
			rating = &ratingVal
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}
		if strings.TrimSpace(cfg.Data.Token) == "" {
			return errors.New("Please login first")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		resp, err := client.UpdateLibraryEntry(cmd.Context(), mangaID, status, rating)
		if err != nil {
			return handleUpdateError(cmd, err, mangaID)
		}

		if config.Runtime().Quiet {
			cmd.Println(resp.MangaID)
			return nil
		}

		cmd.Println("✓ Library entry updated!")
		title := resp.Title
		if title == "" {
			title = "Unknown"
		}
		cmd.Printf("Manga: %s (%s)\n", title, resp.MangaID)
		cmd.Printf("New Status: %s\n", status)
		if rating == nil {
			cmd.Println("New Rating: Unchanged")
		} else {
			cmd.Printf("New Rating: %d/10\n", *rating)
		}
		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(updateCmd)
	updateCmd.Flags().String("manga-id", "", "Manga identifier")
	updateCmd.Flags().String("status", "", "Updated status")
	updateCmd.Flags().Int("rating", 0, "New rating between 1 and 10")
	updateCmd.MarkFlagRequired("manga-id")
	updateCmd.MarkFlagRequired("status")
}

func handleUpdateError(cmd *cobra.Command, err error, mangaID string) error {
	var apiErr *api.Error
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.Status == http.StatusNotFound:
			cmd.Printf("✗ Cannot update: Manga not found in your library: '%s'\n", mangaID)
			cmd.Println("Use: mangahub library add --manga-id <id> --status reading")
			return nil
		case apiErr.Message != "":
			cmd.Printf("✗ Update failed: %s\n", apiErr.Message)
			return nil
		case apiErr.Code != "":
			cmd.Printf("✗ Update failed: %s\n", apiErr.Code)
			return nil
		}
	}
	return err
}
