package stats

import (
	"fmt"
	"strconv"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

var detailedCmd = &cobra.Command{
	Use:     "detailed",
	Short:   "Show detailed statistics",
	Long:    "Display detailed MangaHub statistics including genre breakdowns and activity timelines.",
	Example: "mangahub stats detailed",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		client := api.NewClient(cfg.Data.BaseURL, cfg.Data.Token)
		period, _ := cmd.Flags().GetString("time-period")
		includeGoals, _ := cmd.Flags().GetBool("include-goals")
		var yearPtr, monthPtr *int
		if cmd.Flags().Changed("year") {
			y, _ := cmd.Flags().GetInt("year")
			yearPtr = &y
		}
		if cmd.Flags().Changed("month") {
			m, _ := cmd.Flags().GetInt("month")
			monthPtr = &m
		}

		req := api.ReadingAnalyticsRequest{TimePeriod: period, Year: yearPtr, Month: monthPtr, IncludeGoals: includeGoals}
		stats, err := client.GetReadingAnalytics(cmd.Context(), req)
		if err != nil {
			return err
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, stats)
			return nil
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, stats)
			return nil
		}

		if config.Runtime().Quiet {
			return nil
		}

		cmd.Println("Detailed Reading Statistics")
		cmd.Println()

		if len(stats.FavoriteGenres) > 0 {
			cmd.Println("By Genre:")
			table := utils.Table{Headers: []string{"Genre", "Manga", "Chapters"}}
			for _, genre := range stats.FavoriteGenres {
				table.AddRow(genre.Genre, fmt.Sprintf("%d", genre.Count), fmt.Sprintf("%d", genre.Chapters))
			}
			cmd.Println(table.Render())
			cmd.Println()
		}

		if len(stats.MonthlyStats) > 0 {
			cmd.Println("Recent Monthly Activity:")
			table := utils.Table{Headers: []string{"Month", "Chapters", "Completed", "Started"}}
			for _, m := range stats.MonthlyStats {
				month := fmt.Sprintf("%04d-%02d", m.Year, m.Month)
				table.AddRow(month, fmt.Sprintf("%d", m.ChaptersRead), fmt.Sprintf("%d", m.MangaCompleted), fmt.Sprintf("%d", m.MangaStarted))
			}
			cmd.Println(table.Render())
			cmd.Println()
		}

		if len(stats.YearlyStats) > 0 {
			cmd.Println("Yearly Trends:")
			table := utils.Table{Headers: []string{"Year", "Chapters", "Completed", "Started", "Active Days"}}
			for _, y := range stats.YearlyStats {
				table.AddRow(strconv.Itoa(y.Year), fmt.Sprintf("%d", y.ChaptersRead), fmt.Sprintf("%d", y.MangaCompleted), fmt.Sprintf("%d", y.MangaStarted), fmt.Sprintf("%d", y.TotalDays))
			}
			cmd.Println(table.Render())
			cmd.Println()
		}

		if len(stats.Goals) > 0 {
			cmd.Println("Reading Goals:")
			table := utils.Table{Headers: []string{"Goal", "Progress", "Target", "Status", "Window"}}
			for _, g := range stats.Goals {
				status := "In Progress"
				if g.Completed {
					status = "Completed"
				}
				window := fmt.Sprintf("%s → %s", g.StartDate, g.EndDate)
				table.AddRow(g.GoalType, fmt.Sprintf("%d", g.CurrentValue), fmt.Sprintf("%d", g.TargetValue), status, window)
			}
			cmd.Println(table.Render())
		}

		return nil
	},
}

func init() {
	StatsCmd.AddCommand(detailedCmd)
	output.AddFlag(detailedCmd)
	detailedCmd.Flags().String("time-period", "", "Analytics time period (all_time|year|month|week)")
	detailedCmd.Flags().Int("year", 0, "Filter analytics by year")
	detailedCmd.Flags().Int("month", 0, "Filter analytics by month")
	detailedCmd.Flags().Bool("include-goals", false, "Include reading goals progress")
}
