package stats

import (
	"fmt"

	"github.com/ngocan-dev/mangahub_/cli/internal/api"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/utils"
	"github.com/spf13/cobra"
)

// StatsCmd handles statistics commands.
var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View MangaHub statistics",
	Long:  "Display overview and detailed statistics about MangaHub usage.",
	RunE:  runDateRange,
}

var overviewCmd = &cobra.Command{
	Use:     "overview",
	Short:   "Show overview statistics",
	Long:    "Display summary statistics for your MangaHub library and usage.",
	Example: "mangahub stats overview",
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
		force, _ := cmd.Flags().GetBool("force")
		stats, err := client.GetReadingStatistics(cmd.Context(), force)
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

		cmd.Println("Reading Statistics Overview")
		cmd.Println()
		cmd.Printf("Total Manga Read: %d\n", stats.TotalMangaRead)
		cmd.Printf("Currently Reading: %d\n", stats.TotalMangaReading)
		cmd.Printf("Planned: %d\n", stats.TotalMangaPlanned)
		cmd.Printf("Total Chapters Read: %d\n", stats.TotalChaptersRead)
		cmd.Printf("Average Rating: %.2f\n", stats.AverageRating)
		cmd.Printf("Reading Streak: %d days (longest %d)\n", stats.CurrentStreakDays, stats.LongestStreakDays)
		if stats.TotalReadingTimeHours > 0 {
			cmd.Printf("Time Spent Reading: %.1f hours\n", stats.TotalReadingTimeHours)
		}

		if len(stats.FavoriteGenres) > 0 {
			cmd.Println()
			cmd.Println("Top Genres:")
			table := utils.Table{Headers: []string{"Genre", "Manga", "Chapters"}}
			for _, g := range stats.FavoriteGenres {
				table.AddRow(g.Genre, fmt.Sprintf("%d", g.Count), fmt.Sprintf("%d", g.Chapters))
			}
			cmd.Println(table.Render())
		}

		if len(stats.YearlyStats) > 0 {
			cmd.Println()
			cmd.Println("Yearly Summary:")
			table := utils.Table{Headers: []string{"Year", "Chapters", "Completed", "Started", "Active Days"}}
			for _, y := range stats.YearlyStats {
				table.AddRow(fmt.Sprintf("%d", y.Year), fmt.Sprintf("%d", y.ChaptersRead), fmt.Sprintf("%d", y.MangaCompleted), fmt.Sprintf("%d", y.MangaStarted), fmt.Sprintf("%d", y.TotalDays))
			}
			cmd.Println(table.Render())
		}

		cmd.Println()
		cmd.Printf("Last Calculated: %s\n", stats.LastCalculatedAt)
		return nil
	},
}

func init() {
	StatsCmd.AddCommand(overviewCmd)
	output.AddFlag(overviewCmd)
	overviewCmd.Flags().Bool("force", false, "Force recalculation of statistics")
}
