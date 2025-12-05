package stats

import "github.com/spf13/cobra"

var detailedCmd = &cobra.Command{
	Use:     "detailed",
	Short:   "Show detailed statistics",
	Long:    "Display detailed MangaHub statistics including genre breakdowns and activity timelines.",
	Example: "mangahub stats detailed",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement detailed stats
		cmd.Println("Detailed stats are not yet implemented.")
		return nil
	},
}

func init() {
	StatsCmd.AddCommand(detailedCmd)
}
