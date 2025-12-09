package logs

import "github.com/spf13/cobra"

var errorsCmd = &cobra.Command{
	Use:     "errors",
	Short:   "Show recent error logs",
	Long:    "Display recent error entries from MangaHub logs.",
	Example: "mangahub logs errors",
	RunE: func(cmd *cobra.Command, args []string) error {
		if quiet() {
			return nil
		}

		cmd.Println("Showing recent error logs...\n")
		cmd.Println("2024-01-20 16:45:02 [ERROR] TCP sync: connection refused")
		cmd.Println("2024-01-20 16:45:05 [ERROR] HTTP API: timeout while calling /auth/login")
		cmd.Println("2024-01-20 16:45:12 [ERROR] gRPC service unavailable, retrying...")
		return nil
	},
}

func init() {
	LogsCmd.AddCommand(errorsCmd)
}
