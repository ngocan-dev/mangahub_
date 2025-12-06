package library

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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
		ratingVal, _ := cmd.Flags().GetInt("rating")

		if mangaID == "" {
			return errors.New("--manga-id is required")
		}

		status = strings.ToLower(strings.TrimSpace(status))
		if err := validateStatus(status); err != nil {
			return err
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
		resp, err := client.AddToLibrary(cmd.Context(), mangaID, status, rating)
		if err != nil {
			return handleAddError(cmd, err, mangaID)
		}

		if config.Runtime().Quiet {
			cmd.Println(resp.MangaID)
			return nil
		}

		output.PrintJSON(cmd, resp)
		cmd.Println("✓ Added to your library!")
		title := resp.Title
		if title == "" {
			title = "Unknown"
		}
		cmd.Printf("Manga: %s (%s)\n", title, resp.MangaID)
		cmd.Printf("Status: %s\n", status)
		if rating == nil {
			cmd.Println("Rating: Unrated")
		} else {
			cmd.Printf("Rating: %d/10\n", *rating)
		}
		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(addCmd)
	addCmd.Flags().String("manga-id", "", "Manga identifier")
	addCmd.Flags().String("status", "reading", "Reading status (reading|completed|plan-to-read|on-hold|dropped)")
	addCmd.Flags().Int("rating", 0, "Rating between 1 and 10")
	addCmd.MarkFlagRequired("manga-id")
}

func handleAddError(cmd *cobra.Command, err error, mangaID string) error {
	var apiErr *api.Error
	if errors.As(err, &apiErr) {
		switch {
		case apiErr.Status == http.StatusNotFound:
			cmd.Printf("✗ Cannot add: Manga not found: '%s'\n", mangaID)
			cmd.Println("Try searching first:")
			cmd.Println("mangahub manga search \"title\"")
			return nil
		case strings.Contains(strings.ToLower(apiErr.Message), "already"):
			cmd.Println("✗ This manga is already in your library.")
			cmd.Println("Use: mangahub library update --manga-id <id>")
			return nil
		}
	}
	return err
}

func validateStatus(status string) error {
	switch status {
	case "reading", "completed", "plan-to-read", "on-hold", "dropped":
		return nil
	default:
		return fmt.Errorf("--status must be one of: reading, completed, plan-to-read, on-hold, dropped")
	}
}
