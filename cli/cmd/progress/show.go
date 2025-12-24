package progress

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show reading progress",
	Long:    "Display current reading progress for your library entries.",
	Example: "mangahub progress show",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

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

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{"results": records})
			return nil
		}

		if config.Runtime().Quiet {
			for _, r := range records {
				cmd.Printf("%s,%d\n", r.MangaID, r.Chapter)
			}
			return nil
		}

		if len(records) == 0 {
			cmd.Println("No progress history found.")
			return nil
		}

		if mangaID != "" {
			title := fmt.Sprintf("Reading Progress (%s)", records[0].Manga)
			table := utils.Table{Headers: []string{"Date", "Chapter", "Volume", "Notes", "Source"}}
			for _, r := range records {
				table.AddRow(r.Date.Format("2006-01-02"), fmt.Sprintf("%d", r.Chapter), volumeString(r.Volume), r.Notes, r.Source)
			}
			cmd.Println(table.RenderWithTitle(title))
			return nil
		}

		table := utils.Table{Headers: []string{"Manga", "Date", "Chapter", "Volume", "Source"}}
		for _, r := range records {
			table.AddRow(r.Manga, r.Date.Format("2006-01-02"), fmt.Sprintf("%d", r.Chapter), volumeString(r.Volume), r.Source)
		}

		cmd.Println(table.RenderWithTitle("Reading Progress"))
		return nil
	},
}

func init() {
	ProgressCmd.AddCommand(showCmd)
	showCmd.Flags().String("manga-id", "", "Manga identifier")
	output.AddFlag(showCmd)
}
