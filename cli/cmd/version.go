package cmd

import "github.com/spf13/cobra"

const cliVersion = "0.1.0"

// versionCmd shows CLI version information.
var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Show the MangaHub CLI version",
	Long:    "Display detailed version information for the MangaHub CLI.",
	Example: "mangahub version",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("MangaHub CLI version %s\n", cliVersion)
	},
}
