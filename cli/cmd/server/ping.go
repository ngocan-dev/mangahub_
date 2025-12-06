package server

import "github.com/spf13/cobra"

var pingCmd = &cobra.Command{
	Use:     "ping",
	Short:   "Ping the server",
	Long:    "Send a ping to verify the MangaHub server is reachable.",
	Example: "mangahub server ping",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement server ping
		cmd.Println("Server ping is not yet implemented.")
		return nil
	},
}

func init() {
	ServerCmd.AddCommand(pingCmd)
}
