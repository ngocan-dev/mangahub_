package logs

import "github.com/spf13/cobra"

// LogsCmd groups log inspection commands.
var LogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View and manage MangaHub logs",
	Long:  "Inspect, search, rotate, and clean MangaHub logs.",
}
