package server

import "github.com/spf13/cobra"

var healthCmd = &cobra.Command{
	Use:     "health",
	Short:   "Check server health",
	Long:    "Perform a health check against the MangaHub server.",
	Example: "mangahub server health",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement server health check
		cmd.Println("Server health check is not yet implemented.")
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(healthCmd)
}
