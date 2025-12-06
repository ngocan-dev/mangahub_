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

var removeCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove a manga from the library",
	Long:    "Delete a manga from your library using its identifier.",
	Example: "mangahub library remove --manga-id one-piece",
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		if strings.TrimSpace(mangaID) == "" {
			return errors.New("--manga-id is required")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}
		if strings.TrimSpace(cfg.Data.Token) == "" {
			return errors.New("Please login first")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		if err := client.RemoveFromLibrary(cmd.Context(), mangaID); err != nil {
			var apiErr *api.Error
			if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
				cmd.Printf("✗ Cannot remove: Manga not found in your library: '%s'\n", mangaID)
				return nil
			}
			return err
		}

		if config.Runtime().Quiet {
			cmd.Println(mangaID)
			return nil
		}

		cmd.Println("✓ Removed from your library.")
		cmd.Printf("Manga: %s\n", mangaID)
		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(removeCmd)
	removeCmd.Flags().String("manga-id", "", "Manga identifier")
	removeCmd.MarkFlagRequired("manga-id")
}
