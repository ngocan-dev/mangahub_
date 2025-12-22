package novel

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search <keyword>",
	Short:   "Search for novels",
	Long:    "Search MangaHub for novels using keywords.",
	Example: "mangahub novel search \"one piece\"",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		keyword := strings.TrimSpace(args[0])
		if keyword == "" {
			return errors.New("keyword cannot be empty")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		resp, err := client.SearchManga(cmd.Context(), keyword, api.MangaSearchFilters{})
		if err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{
				"results": resp.Results,
				"total":   resp.Total,
				"page":    resp.Page,
				"limit":   resp.Limit,
				"pages":   resp.Pages,
			})
			return nil
		}

		if config.Runtime().Quiet {
			for _, res := range resp.Results {
				cmd.Println(res.ID)
			}
			return nil
		}

		cmd.Printf("Searching novels for \"%s\"...\n", keyword)
		if len(resp.Results) == 0 {
			cmd.Println("No novels found matching your search criteria.")
			return nil
		}

		table := utils.Table{Headers: []string{"ID", "Title", "Author", "Status", "Genre"}}
		for _, novel := range resp.Results {
			table.AddRow(
				fmt.Sprintf("%d", novel.ID),
				formatNovelTitle(novel.Title, novel.Name),
				novel.Author,
				formatNovelStatus(novel.Status),
				novel.Genre,
			)
		}

		cmd.Println(table.RenderWithTitle("Novel Search Results"))
		cmd.Println("Use 'mangahub novel info <novelID>' to view details")
		return nil
	},
}

func init() {
	NovelCmd.AddCommand(searchCmd)
	output.AddFlag(searchCmd)
}

func formatNovelTitle(title, name string) string {
	title = strings.TrimSpace(title)
	name = strings.TrimSpace(name)
	if title == "" {
		return name
	}
	if name == "" || strings.EqualFold(name, title) {
		return title
	}
	return fmt.Sprintf("%s (%s)", title, name)
}

func formatNovelStatus(status string) string {
	lower := strings.ToLower(status)
	switch lower {
	case "ongoing":
		return "Ongoing"
	case "completed":
		return "Completed"
	case "hiatus":
		return "Hiatus"
	case "canceled":
		return "Canceled"
	default:
		return status
	}
}
