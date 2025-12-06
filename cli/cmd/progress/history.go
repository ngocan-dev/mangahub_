package progress

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:     "history",
	Short:   "Show reading history",
	Long:    "Display historical reading activity for your library entries.",
	Example: "mangahub progress history --manga-id one-piece",
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")

		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}
		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)

		records, err := client.ProgressHistory(cmd.Context(), mangaID)
		if err != nil {
			return err
		}

		output.PrintJSON(cmd, records)
		if config.Runtime().Quiet {
			return nil
		}

		if mangaID != "" {
			if len(records) == 0 {
				cmd.Println("No progress history found.")
				return nil
			}
			title := fmt.Sprintf("Reading Progress History (%s)", records[0].Manga)
			table := utils.Table{Headers: []string{"Date", "Chapter", "Volume", "Notes", "Source"}}
			for _, r := range records {
				table.AddRow(r.Date.Format("2006-01-02"), fmt.Sprintf("%d", r.Chapter), volumeString(r.Volume), r.Notes, r.Source)
			}
			cmd.Println(table.RenderWithTitle(title))
			return nil
		}

		if len(records) == 0 {
			cmd.Println("No progress history found.")
			return nil
		}

		table := utils.Table{Headers: []string{"Manga", "Date", "Chapter", "Volume", "Source"}}
		for _, r := range records {
			table.AddRow(r.Manga, r.Date.Format("2006-01-02"), fmt.Sprintf("%d", r.Chapter), volumeString(r.Volume), r.Source)
		}

		cmd.Println(table.RenderWithTitle("Reading Progress History"))
		return nil
	},
}

func volumeString(v int) string {
	if v <= 0 {
		return ""
	}
	return fmt.Sprintf("%d", v)
}

func init() {
	ProgressCmd.AddCommand(historyCmd)
	historyCmd.Flags().String("manga-id", "", "Manga identifier")
}
