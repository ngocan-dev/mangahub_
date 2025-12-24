package manga

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List manga",
	Long:    "List manga available in MangaHub with optional pagination.",
	Example: "mangahub manga list --page 1 --limit 20",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		filters, err := parseListFilters(cmd)
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		items, err := client.ListManga(cmd.Context(), filters)
		if err != nil {
			return err
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, items)
			return nil
		}

		if config.Runtime().Quiet {
			for _, item := range items {
				cmd.Println(item.ID)
			}
			return nil
		}

		cmd.Printf("Listing manga (page %d, limit %d)...\n\n", filters.Page, filters.Limit)
		if len(items) == 0 {
			cmd.Println("No manga available for your selected filters.")
			cmd.Println("Try adjusting filters or search directly:")
			cmd.Println("mangahub manga search \"<keyword>\"")
			return nil
		}

		rows := fromListItems(items)
		table := buildMangaTable(rows)
		cmd.Print(table)
		return nil
	},
}

func init() {
	MangaCmd.AddCommand(listCmd)
	listCmd.Flags().Int("page", 1, "Page number")
	listCmd.Flags().Int("limit", 20, "Results per page")
	listCmd.Flags().String("genre", "", "Filter by genre")
	listCmd.Flags().String("status", "", "Filter by status (ongoing|completed|hiatus|canceled)")
}

func parseListFilters(cmd *cobra.Command) (api.MangaListFilters, error) {
	page, _ := cmd.Flags().GetInt("page")
	limit, _ := cmd.Flags().GetInt("limit")
	genre, _ := cmd.Flags().GetString("genre")
	status, _ := cmd.Flags().GetString("status")

	if page < 1 {
		return api.MangaListFilters{}, errors.New("--page must be greater than 0")
	}
	if limit < 1 {
		return api.MangaListFilters{}, errors.New("--limit must be greater than 0")
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" {
		switch status {
		case "ongoing", "completed", "hiatus", "canceled":
		default:
			return api.MangaListFilters{}, errors.New("--status must be one of: ongoing, completed, hiatus, canceled")
		}
	}

	return api.MangaListFilters{
		Page:   page,
		Limit:  limit,
		Genre:  strings.TrimSpace(genre),
		Status: status,
	}, nil
}

func fromListItems(items []api.MangaListItem) [][]string {
	var rows [][]string
	for _, item := range items {
		rows = append(rows, []string{
			item.ID,
			formatTitleCell(item.Title, strings.Join(item.AltTitles, ", ")),
			formatAuthorCell(item.Author),
			formatStatus(item.Status),
			formatChapterCount(item.Chapters),
		})
	}
	return rows
}

func formatChapterCount(chapters int) string {
	if chapters <= 0 {
		return "Unknown"
	}
	return fmt.Sprintf("%d", chapters)
}
