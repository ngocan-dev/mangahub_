package fetch

import "github.com/spf13/cobra"

// FetchCmd groups metadata fetch commands.
var FetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch metadata from MangaHub",
	Long:  "Retrieve metadata snapshots from MangaHub services.",
}
