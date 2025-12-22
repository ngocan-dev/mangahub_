package fav

import (
	"errors"
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List favorite novels",
	Long:    "List all novels saved in your favorites/bookmarks.",
	Example: "mangahub fav list",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		entries, err := client.ListLibrary(cmd.Context(), "plan_to_read", "", "")
		if err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{"results": entries})
			return nil
		}

		if config.Runtime().Quiet {
			for _, entry := range entries {
				cmd.Println(entry.MangaID)
			}
			return nil
		}

		if len(entries) == 0 {
			cmd.Println("No favorites yet. Add one with: mangahub fav add <novelID>")
			return nil
		}

		table := utils.Table{Headers: []string{"ID", "Title", "Status", "Current Chapter"}}
		for _, entry := range entries {
			chapter := fmt.Sprintf("%d", entry.CurrentChapter)
			table.AddRow(entry.MangaID, entry.Title, entry.Status, chapter)
		}

		cmd.Println(table.RenderWithTitle("Favorite Novels"))
		return nil
	},
}

func init() {
	FavCmd.AddCommand(listCmd)
	output.AddFlag(listCmd)
}
