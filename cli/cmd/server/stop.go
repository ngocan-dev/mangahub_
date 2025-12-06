package server

import "github.com/spf13/cobra"

var stopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop the server",
	Long:    "Stop the running MangaHub server instance.",
	Example: "mangahub server stop",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement server stop logic
		cmd.Println("Server stop is not yet implemented.")
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(stopCmd)
}
