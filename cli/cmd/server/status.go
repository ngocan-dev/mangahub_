package server

import "github.com/spf13/cobra"

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show server status",
	Long:    "Display the current status of the MangaHub server.",
	Example: "mangahub server status",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement server status logic
		cmd.Println("Server status is not yet implemented.")
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(statusCmd)
}
