package manga

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

// MangaCmd groups manga-related commands.
var MangaCmd = &cobra.Command{
	Use:   "manga",
	Short: "Interact with manga metadata",
	Long:  "Search and retrieve manga information from MangaHub services.",
}

// searchCmd performs manga searches against the backend service.
var searchCmd = &cobra.Command{
	Use:     "search <query>",
	Short:   "Search for manga titles",
	Long:    "Search MangaHub for manga titles using keywords.",
	Example: "mangahub manga search \"one piece\"",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		query := args[0]

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		filters, err := parseSearchFilters(cmd)
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		results, err := client.SearchManga(cmd.Context(), query, filters)
		if err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{"results": results})
			return nil
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, results)
			return nil
		}

		if config.Runtime().Quiet {
			for _, res := range results {
				cmd.Println(res.ID)
			}
			return nil
		}

		cmd.Printf("Searching for \"%s\"...\n", query)
		if len(results) == 0 {
			cmd.Println("No manga found matching your search criteria.")
			cmd.Println("\nSuggestions:")
			cmd.Println("- Check spelling and try again")
			cmd.Println("- Use broader search terms")
			cmd.Println("- Browse by genre: mangahub manga list --genre action")
			return nil
		}

		cmd.Printf("Found %d results:\n", len(results))
		table := buildMangaTable(fromSearchResults(results))
		cmd.Print(table)
		cmd.Println("\nUse 'mangahub manga info <id>' to view details")
		cmd.Println("Use 'mangahub library add --manga-id <id>' to add to your library")
		return nil
	},
}

func init() {
	MangaCmd.AddCommand(searchCmd)
	searchCmd.Flags().String("genre", "", "Filter by genre (comma-separated)")
	searchCmd.Flags().String("status", "", "Filter by status (ongoing|completed|hiatus|canceled)")
	searchCmd.Flags().String("author", "", "Filter by author")
	searchCmd.Flags().Int("year-from", 0, "Filter by starting publication year")
	searchCmd.Flags().Int("year-to", 0, "Filter by ending publication year")
	searchCmd.Flags().Int("min-chapters", 0, "Filter by minimum chapter count")
	searchCmd.Flags().String("sort-by", "", "Sort by field (title|chapters|year|popularity)")
	searchCmd.Flags().String("order", "", "Sort order (asc|desc)")
	searchCmd.Flags().Int("limit", 0, "Limit number of results")
	output.AddFlag(searchCmd)
}

func parseSearchFilters(cmd *cobra.Command) (api.MangaSearchFilters, error) {
	status, _ := cmd.Flags().GetString("status")
	sortBy, _ := cmd.Flags().GetString("sort-by")
	order, _ := cmd.Flags().GetString("order")
	genre, _ := cmd.Flags().GetString("genre")
	author, _ := cmd.Flags().GetString("author")
	yearFrom, _ := cmd.Flags().GetInt("year-from")
	yearTo, _ := cmd.Flags().GetInt("year-to")
	minChapters, _ := cmd.Flags().GetInt("min-chapters")
	limit, _ := cmd.Flags().GetInt("limit")

	if yearFrom > 0 && yearTo > 0 && yearFrom > yearTo {
		return api.MangaSearchFilters{}, errors.New("--year-from cannot be greater than --year-to")
	}

	if status != "" {
		status = strings.ToLower(status)
		switch status {
		case "ongoing", "completed", "hiatus", "canceled":
		default:
			return api.MangaSearchFilters{}, errors.New("--status must be one of: ongoing, completed, hiatus, canceled")
		}
	}

	if sortBy != "" {
		sortBy = strings.ToLower(sortBy)
		switch sortBy {
		case "title", "chapters", "year", "popularity":
		default:
			return api.MangaSearchFilters{}, errors.New("--sort-by must be one of: title, chapters, year, popularity")
		}
	}

	if order != "" {
		order = strings.ToLower(order)
		switch order {
		case "asc", "desc":
		default:
			return api.MangaSearchFilters{}, errors.New("--order must be one of: asc, desc")
		}
	}

	return api.MangaSearchFilters{
		Genre:       strings.TrimSpace(genre),
		Status:      status,
		Author:      strings.TrimSpace(author),
		YearFrom:    yearFrom,
		YearTo:      yearTo,
		MinChapters: minChapters,
		SortBy:      sortBy,
		Order:       order,
		Limit:       limit,
	}, nil
}

func buildMangaTable(rows [][]string) string {
	t := utils.Table{Headers: []string{"ID", "Title", "Author", "Status", "Chapters"}, Rows: rows}
	return t.Render()
}

func fromSearchResults(results []api.MangaSearchResult) [][]string {
	var rows [][]string
	for _, res := range results {
		rows = append(rows, []string{
			res.ID,
			formatTitleCell(res.Title, res.AltTitles),
			formatAuthorCell(res.Author),
			formatStatus(res.Status),
			formatChapterCount(res.Chapters),
		})
	}
	return rows
}

func formatTitleCell(title string, alt []string) string {
	if len(alt) == 0 {
		return title
	}
	altTitle := alt[0]
	if altTitle != "" && !strings.HasPrefix(altTitle, "(") {
		altTitle = fmt.Sprintf("(%s)", altTitle)
	}
	return strings.Join([]string{title, altTitle}, "\n")
}

func formatAuthorCell(author string) string {
	parts := strings.Split(strings.TrimSpace(author), " ")
	if len(parts) <= 1 {
		return author
	}
	return strings.Join(parts, "\n")
}

func formatStatus(status string) string {
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

func formatChapterCount(chapters int) string {
	if chapters <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d", chapters)
}
