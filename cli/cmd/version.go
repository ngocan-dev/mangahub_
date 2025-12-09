package cmd

import (
	"runtime"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	"github.com/ngocan-dev/mangahub_/cli/internal/version"
	"github.com/spf13/cobra"
)

// versionCmd shows CLI version information.
var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Show the MangaHub CLI version",
	Long:    "Display detailed version information for the MangaHub CLI, including build metadata and backend compatibility.",
	Example: "mangahub version",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		info := map[string]string{
			"version":    version.CLIVersion,
			"build_time": version.BuildTime,
			"go_version": runtime.Version(),
			"api_compat": version.APICompatibility,
		}

		if format == output.FormatJSON {
			output.PrintJSON(cmd, info)
			return nil
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, info)
			return nil
		}

		if config.Runtime().Quiet {
			return nil
		}

		cmd.Printf("MangaHub CLI version: %s\n", version.CLIVersion)
		cmd.Printf("Build: %s\n", version.BuildTime)
		cmd.Printf("Go version: %s\n", runtime.Version())
		cmd.Printf("Backend API version: %s compatible\n", version.APICompatibility)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	output.AddFlag(versionCmd)
}
