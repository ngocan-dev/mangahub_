package stats

import (
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
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

		details := map[string]any{
			"genres": []map[string]any{
				{"name": "Shounen", "manga": 12, "chapters": 3420},
				{"name": "Seinen", "manga": 6, "chapters": 1110},
				{"name": "Romance", "manga": 4, "chapters": 230},
				{"name": "Comedy", "manga": 3, "chapters": 580},
			},
			"status": map[string]int{
				"completed": 28,
				"reading":   8,
				"on_hold":   3,
				"dropped":   3,
			},
			"top_manga": []map[string]any{
				{"title": "One Piece", "chapters_read": 1095, "time_spent": "52h 20m", "rating": "9/10"},
				{"title": "Naruto", "chapters_read": 700, "time_spent": "33h 00m", "rating": "8/10"},
				{"title": "Attack on Titan", "chapters_read": 139, "time_spent": "11h 12m", "rating": "Unrated"},
			},
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, details)
			return nil
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, details)
			return nil
		}

		if config.Runtime().Quiet {
			return nil
		}

		cmd.Println("Detailed Reading Statistics")
		cmd.Println()
		cmd.Println("By Genre:")
		cmd.Println("• Shounen: 12 manga (3,420 chapters)")
		cmd.Println("• Seinen: 6 manga (1,110 chapters)")
		cmd.Println("• Romance: 4 manga (230 chapters)")
		cmd.Println("• Comedy: 3 manga (580 chapters)")
		cmd.Println()
		cmd.Println("By Status:")
		cmd.Println("• Completed: 28")
		cmd.Println("• Reading: 8")
		cmd.Println("• On-Hold: 3")
		cmd.Println("• Dropped: 3")
		cmd.Println()
		cmd.Println("Top 10 Read Manga:")
		cmd.Println("┌───────────────────────┬──────────────┬─────────────┬──────────┐")
		cmd.Println("│ Title                 │ Chapters Read│ Time Spent  │ Rating   │")
		cmd.Println("├───────────────────────┼──────────────┼─────────────┼──────────┤")
		cmd.Println("│ One Piece             │ 1,095        │ 52h 20m     │ 9/10     │")
		cmd.Println("│ Naruto                │ 700          │ 33h 00m     │ 8/10     │")
		cmd.Println("│ Attack on Titan       │ 139          │ 11h 12m     │ Unrated  │")
		cmd.Println("└───────────────────────┴──────────────┴─────────────┴──────────┘")
		return nil
	},
}

func init() {
	StatsCmd.AddCommand(detailedCmd)
	output.AddFlag(detailedCmd)
}
