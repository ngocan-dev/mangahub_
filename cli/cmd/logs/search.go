package logs

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search",
	Short:   "Search logs",
	Long:    "Search MangaHub logs for a specific phrase.",
	Example: "mangahub logs search \"connection failed\"",
	RunE: func(cmd *cobra.Command, args []string) error {
		if quiet() {
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("search term is required")
		}

		query := args[0]
		cmd.Printf("Searching logs for: \"%s\"\n\n", query)
		cmd.Println("2024-01-20 10:15:22 [ERROR] connection failed: no route to host")
		cmd.Println("2024-01-20 11:47:03 [WARN] connection failed, retry in 5s")
		return nil
	},
}

func init() {
	LogsCmd.AddCommand(searchCmd)
}
