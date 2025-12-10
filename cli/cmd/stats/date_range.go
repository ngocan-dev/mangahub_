package stats

import (
	"fmt"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

func runDateRange(cmd *cobra.Command, args []string) error {
	format, err := output.GetFormat(cmd)
	if err != nil {
		return err
	}

	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")

	if from != "" || to != "" {
		if from == "" || to == "" {
			return fmt.Errorf("both --from and --to must be provided")
		}

		fromDate, err := time.Parse("2006-01-02", from)
		if err != nil {
			return fmt.Errorf("invalid --from date: %w", err)
		}
		toDate, err := time.Parse("2006-01-02", to)
		if err != nil {
			return fmt.Errorf("invalid --to date: %w", err)
		}
		if toDate.Before(fromDate) {
			return fmt.Errorf("--to must be on or after --from")
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		stats, err := client.GetReadingStatistics(cmd.Context(), false)
		if err != nil {
			return err
		}

		summary := filterStatsByRange(stats, fromDate, toDate)

		if format == output.FormatJSON {
			output.PrintJSON(cmd, summary)
			return nil
		}

		if config.Runtime().Quiet {
			return nil
		}
		cmd.Println(fmt.Sprintf("Reading Statistics (%s → %s)", from, to))
		cmd.Println()
		cmd.Printf("Total Chapters Read: %d\n", summary.TotalChapters)
		cmd.Printf("Manga Completed: %d\n", summary.MangaCompleted)
		cmd.Printf("Manga Started: %d\n", summary.MangaStarted)
		if summary.Months > 0 {
			cmd.Printf("Average Chapters per Month: %.1f\n", summary.AveragePerMonth)
		}

		if len(summary.MonthlyStats) > 0 {
			cmd.Println()
			cmd.Println("Monthly Breakdown:")
			table := utils.Table{Headers: []string{"Month", "Chapters", "Completed", "Started"}}
			for _, m := range summary.MonthlyStats {
				month := fmt.Sprintf("%04d-%02d", m.Year, m.Month)
				table.AddRow(month, fmt.Sprintf("%d", m.ChaptersRead), fmt.Sprintf("%d", m.MangaCompleted), fmt.Sprintf("%d", m.MangaStarted))
			}
			cmd.Println(table.Render())
		}

		return nil
	}

	return cmd.Help()
}

func init() {
	StatsCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	StatsCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")
	output.AddFlag(StatsCmd)
}

type rangeSummary struct {
	Range           map[string]string `json:"range"`
	TotalChapters   int               `json:"total_chapters"`
	MangaCompleted  int               `json:"manga_completed"`
	MangaStarted    int               `json:"manga_started"`
	Months          int               `json:"months"`
	AveragePerMonth float64           `json:"average_per_month"`
	MonthlyStats    []api.MonthlyStat `json:"monthly_stats"`
}

func filterStatsByRange(stats *api.ReadingStatistics, from, to time.Time) rangeSummary {
	result := rangeSummary{
		Range: map[string]string{
			"from": from.Format("2006-01-02"),
			"to":   to.Format("2006-01-02"),
		},
	}

	if stats == nil {
		return result
	}

	startMonth := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
	endMonth := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, time.UTC)

	for _, m := range stats.MonthlyStats {
		monthTime := time.Date(m.Year, time.Month(m.Month), 1, 0, 0, 0, 0, time.UTC)
		if monthTime.Before(startMonth) || monthTime.After(endMonth) {
			continue
		}

		result.MonthlyStats = append(result.MonthlyStats, m)
		result.TotalChapters += m.ChaptersRead
		result.MangaCompleted += m.MangaCompleted
		result.MangaStarted += m.MangaStarted
	}

	result.Months = len(result.MonthlyStats)
	if result.Months > 0 {
		result.AveragePerMonth = float64(result.TotalChapters) / float64(result.Months)
	}

	return result
}
