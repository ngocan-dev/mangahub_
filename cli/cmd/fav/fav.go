package fav

import "github.com/spf13/cobra"

// FavCmd manages favorite/bookmark commands.
var FavCmd = &cobra.Command{
	Use:   "fav",
	Short: "Manage favorite novels",
	Long:  "Add or remove favorites and list your bookmarked novels.",
}
