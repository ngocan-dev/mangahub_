package stats

import "github.com/spf13/cobra"

var detailedCmd = &cobra.Command{
	Use:     "detailed",
	Short:   "Show detailed statistics",
	Long:    "Display detailed MangaHub statistics including genre breakdowns and activity timelines.",
	Example: "mangahub stats detailed",
	RunE: func(cmd *cobra.Command, args []string) error {
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
}
