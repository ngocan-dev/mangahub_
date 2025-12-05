package server

import "github.com/spf13/cobra"

// ServerCmd controls the MangaHub server lifecycle.
var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the MangaHub server",
	Long:  "Start, stop, and monitor the MangaHub server.",
}

var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start the server",
	Long:    "Start the MangaHub local server instance.",
	Example: "mangahub server start --port 8080",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement server start logic
		cmd.Println("Server start is not yet implemented.")
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(startCmd)
	startCmd.Flags().Int("port", 8080, "Port to run the server on")
}
