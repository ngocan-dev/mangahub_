package library

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List library entries",
	Long:    "List all manga saved in your MangaHub library.",
	Example: "mangahub library list",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		order, _ := cmd.Flags().GetString("order")

		status = strings.ToLower(strings.TrimSpace(status))
		if status != "" {
			if err := validateStatus(status); err != nil {
				cmd.Println("✗ Invalid status. Must be one of:")
				cmd.Println("reading, completed, plan-to-read, on-hold, dropped")
				return nil
			}
		}

		sortBy = strings.ToLower(strings.TrimSpace(sortBy))
		if sortBy != "" {
			switch sortBy {
			case "title", "rating", "chapter", "last-updated":
			default:
				return fmt.Errorf("--sort-by must be one of: title, rating, chapter, last-updated")
			}
		}

		order = strings.ToLower(strings.TrimSpace(order))
		if order == "" {
			order = "asc"
		}
		switch order {
		case "asc", "desc":
		default:
			return fmt.Errorf("--order must be one of: asc, desc")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return errors.New("configuration not loaded")
		}
		if strings.TrimSpace(cfg.Data.Token) == "" {
			return errors.New("Please login first")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		entries, err := client.ListLibrary(cmd.Context(), status, sortBy, order)
		if err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, map[string]any{"results": entries})
			return nil
		}

		if config.Runtime().Verbose {
			printVerbose(cmd, entries)
			return nil
		}

		if config.Runtime().Quiet {
			for _, entry := range entries {
				cmd.Println(entry.MangaID)
			}
			return nil
		}

		if len(entries) == 0 {
			cmd.Println("Your library is empty.")
			cmd.Println()
			cmd.Println("Get started by searching and adding manga:")
			cmd.Println("  mangahub manga search \"your favorite series\"")
			cmd.Println("  mangahub library add --manga-id <id> --status reading")
			return nil
		}

		cmd.Printf("Your Manga Library (%d entries)\n\n", len(entries))
		renderGroupedLibrary(cmd, entries, sortBy, order, status)
		cmd.Println()
		cmd.Println("Use --status <status> to filter by specific status")
		cmd.Println("Use --verbose for detailed view with descriptions")
		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(listCmd)
	listCmd.Flags().String("status", "", "Filter by status (reading|completed|plan-to-read|on-hold|dropped)")
	listCmd.Flags().String("sort-by", "", "Sort by (title|rating|chapter|last-updated)")
	listCmd.Flags().String("order", "asc", "Sort order (asc|desc)")
	output.AddFlag(listCmd)
}

func printVerbose(cmd *cobra.Command, entries []api.LibraryEntry) {
	data, _ := json.MarshalIndent(entries, "", "  ")
	cmd.Println(string(data))
}

func renderGroupedLibrary(cmd *cobra.Command, entries []api.LibraryEntry, sortBy, order, filterStatus string) {
	groups := groupByStatus(entries)

	if filterStatus != "" {
		list := groups[strings.ToLower(filterStatus)]
		if len(list) == 0 {
			cmd.Println("Your library is empty.")
			return
		}
		cmd.Printf("%s (%d):\n", humanizeStatus(filterStatus), len(list))
		if filterStatus == "completed" {
			cmd.Print(renderCompletedTable(list, sortBy, order))
		} else {
			cmd.Print(renderReadingTable(list, sortBy, order))
		}
		return
	}

	if len(groups["reading"]) > 0 {
		cmd.Printf("Currently Reading (%d):\n", len(groups["reading"]))
		cmd.Print(renderReadingTable(groups["reading"], sortBy, order))
		cmd.Println()
	}

	if len(groups["completed"]) > 0 {
		cmd.Printf("Completed (%d):\n", len(groups["completed"]))
		cmd.Print(renderCompletedTable(groups["completed"], sortBy, order))
		cmd.Println()
	}

	var summary []string
	for _, key := range []string{"plan-to-read", "on-hold", "dropped"} {
		if len(groups[key]) > 0 {
			summary = append(summary, fmt.Sprintf("%s (%d)", humanizeStatus(key), len(groups[key])))
		}
	}

	if len(summary) > 0 {
		cmd.Println(strings.Join(summary, ", "))
	}
}

func renderReadingTable(entries []api.LibraryEntry, sortBy, order string) string {
	sortEntries(entries, sortBy, order)

	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, []string{
			e.MangaID,
			e.Title,
			formatChapterProgress(e.CurrentChapter, e.TotalChapters),
			formatRating(e.Rating),
			formatStartedAt(e.StartedAt),
			humanizeUpdatedAt(e.UpdatedAt),
		})
	}
	t := utils.Table{Headers: []string{"ID", "Title", "Chapter", "Rating", "Started", "Updated"}, Rows: rows}
	return t.Render()
}

func renderCompletedTable(entries []api.LibraryEntry, sortBy, order string) string {
	sortEntries(entries, sortBy, order)

	rows := make([][]string, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, []string{
			e.MangaID,
			e.Title,
			formatChapterProgress(e.CurrentChapter, e.TotalChapters),
			formatRating(e.Rating),
			formatCompletedAt(e.CompletedAt, e.UpdatedAt),
		})
	}
	t := utils.Table{Headers: []string{"ID", "Title", "Chapters", "Rating", "Completed"}, Rows: rows}
	return t.Render()
}

func groupByStatus(entries []api.LibraryEntry) map[string][]api.LibraryEntry {
	groups := map[string][]api.LibraryEntry{
		"reading":      {},
		"completed":    {},
		"plan-to-read": {},
		"on-hold":      {},
		"dropped":      {},
	}

	for _, entry := range entries {
		status := strings.ToLower(entry.Status)
		groups[status] = append(groups[status], entry)
	}
	return groups
}

func sortEntries(entries []api.LibraryEntry, sortBy, order string) {
	if sortBy == "" {
		sortBy = "title"
	}

	sort.SliceStable(entries, func(i, j int) bool {
		left, right := entries[i], entries[j]

		switch sortBy {
		case "rating":
			li := -1
			if left.Rating != nil {
				li = *left.Rating
			}
			lj := -1
			if right.Rating != nil {
				lj = *right.Rating
			}
			if li == lj {
				return compareStrings(left.Title, right.Title, order)
			}
			if order == "desc" {
				return li > lj
			}
			return li < lj
		case "chapter":
			if left.CurrentChapter == right.CurrentChapter {
				return compareStrings(left.Title, right.Title, order)
			}
			if order == "desc" {
				return left.CurrentChapter > right.CurrentChapter
			}
			return left.CurrentChapter < right.CurrentChapter
		case "last-updated":
			lt := parseTime(left.UpdatedAt)
			rt := parseTime(right.UpdatedAt)
			if lt.Equal(rt) {
				return compareStrings(left.Title, right.Title, order)
			}
			if order == "desc" {
				return lt.After(rt)
			}
			return lt.Before(rt)
		default: // title
			return compareStrings(left.Title, right.Title, order)
		}
	})
}

func compareStrings(a, b, order string) bool {
	if order == "desc" {
		return strings.ToLower(a) > strings.ToLower(b)
	}
	return strings.ToLower(a) < strings.ToLower(b)
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t
	}
	return time.Time{}
}

func formatChapterProgress(current int, total *int) string {
	totalText := "??"
	if total != nil && *total > 0 {
		totalText = fmt.Sprintf("%d", *total)
	}
	if current <= 0 {
		return fmt.Sprintf("0/%s", totalText)
	}
	return fmt.Sprintf("%d/%s", current, totalText)
}

func formatRating(rating *int) string {
	if rating == nil || *rating <= 0 {
		return "Unrated"
	}
	return fmt.Sprintf("%d/10", *rating)
}

func formatStartedAt(started string) string {
	if started == "" {
		return "-"
	}
	if t, err := time.Parse("2006-01-02", started); err == nil {
		return t.Format("2006-01")
	}
	return started
}

func formatCompletedAt(completed, updated string) string {
	date := completed
	if date == "" {
		date = updated
	}
	if date == "" {
		return "-"
	}
	if t, err := time.Parse(time.RFC3339, date); err == nil {
		return t.Format("2006-01-02")
	}
	if t, err := time.Parse("2006-01-02", date); err == nil {
		return t.Format("2006-01-02")
	}
	return date
}

func humanizeUpdatedAt(updated string) string {
	if updated == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, updated)
	if err != nil {
		return updated
	}
	diff := time.Since(t)
	if diff < 0 {
		diff = 0
	}
	switch {
	case diff < 2*time.Minute:
		return "Just now"
	case diff < time.Hour:
		return fmt.Sprintf("%d mins", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d hours", int(diff.Hours()))
	case diff < 14*24*time.Hour:
		days := int(diff.Hours()) / 24
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	default:
		weeks := int(diff.Hours()) / (24 * 7)
		if weeks <= 1 {
			return "1 week"
		}
		return fmt.Sprintf("%d weeks", weeks)
	}
}

func humanizeStatus(status string) string {
	switch strings.ToLower(status) {
	case "reading":
		return "Currently Reading"
	case "completed":
		return "Completed"
	case "plan-to-read":
		return "Plan to Read"
	case "on-hold":
		return "On Hold"
	case "dropped":
		return "Dropped"
	default:
		return status
	}
}
