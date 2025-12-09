package stats

import (
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
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

		summary := map[string]any{
			"total_manga_read":         42,
			"total_chapters_read":      9832,
			"average_chapters_per_day": 27.4,
			"reading_streak_days":      45,
			"most_active_day":          "2024-01-12",
			"most_active_day_count":    312,
			"favorite_genre":           "Shounen",
			"top_manga": map[string]any{
				"title":         "One Piece",
				"chapters_read": 1095,
			},
			"time_spent_hours": 184,
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, summary)
			return nil
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, summary)
			return nil
		}

		if config.Runtime().Quiet {
			return nil
		}

		cmd.Println("Reading Statistics Overview")
		cmd.Println()
		cmd.Println("Total Manga Read: 42")
		cmd.Println("Total Chapters Read: 9,832")
		cmd.Println("Average Chapters per Day: 27.4")
		cmd.Println("Reading Streak: 45 days")
		cmd.Println("Most Active Day: 2024-01-12 (312 chapters)")
		cmd.Println("Favorite Genre: Shounen")
		cmd.Println("Top Manga: One Piece (1,095 chapters read)")
		cmd.Println("Time Spent Reading: 184 hours")
		return nil
	},
}

func init() {
	StatsCmd.AddCommand(overviewCmd)
	output.AddFlag(overviewCmd)
}
