package fetch

import (
	"errors"
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

var metadataCmd = &cobra.Command{
	Use:     "metadata",
	Short:   "Fetch manga metadata",
	Long:    "Fetch a snapshot of manga metadata from the backend service.",
	Example: "mangahub fetch metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		items, err := client.ListManga(cmd.Context(), api.MangaListFilters{Page: page, Limit: limit})
		if err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{"results": items})
			return nil
		}

		if config.Runtime().Quiet {
			for _, item := range items {
				cmd.Println(item.ID)
			}
			return nil
		}

		if len(items) == 0 {
			cmd.Println("No metadata returned.")
			return nil
		}

		table := utils.Table{Headers: []string{"ID", "Title", "Author", "Status", "Chapters"}}
		for _, item := range items {
			table.AddRow(item.ID, item.Title, item.Author, item.Status, fmt.Sprintf("%d", item.Chapters))
		}

		cmd.Println(table.RenderWithTitle("Metadata Snapshot"))
		return nil
	},
}

func init() {
	FetchCmd.AddCommand(metadataCmd)
	metadataCmd.Flags().Int("page", 1, "Page number")
	metadataCmd.Flags().Int("limit", 20, "Results per page")
	output.AddFlag(metadataCmd)
}
