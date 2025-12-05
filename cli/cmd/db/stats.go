package db

import "github.com/spf13/cobra"

var statsCmd = &cobra.Command{
	Use:     "stats",
	Short:   "Show database stats",
	Long:    "Display database statistics such as size and record counts.",
	Example: "mangahub db stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement db stats
		cmd.Println("Database stats are not yet implemented.")
		return nil
	},
}

func init() {
	DBCmd.AddCommand(statsCmd)
}
