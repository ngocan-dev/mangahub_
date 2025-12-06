package stats

import (
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
}
