package stats

import "github.com/spf13/cobra"

// StatsCmd handles statistics commands.
var StatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View MangaHub statistics",
	Long:  "Display overview and detailed statistics about MangaHub usage.",
}

var overviewCmd = &cobra.Command{
	Use:     "overview",
	Short:   "Show overview statistics",
	Long:    "Display summary statistics for your MangaHub library and usage.",
	Example: "mangahub stats overview",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement stats overview
		cmd.Println("Stats overview is not yet implemented.")
		return nil
	},
}

func init() {
	StatsCmd.AddCommand(overviewCmd)
}
