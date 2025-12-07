package update

import (
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	updateclient "github.com/ngocan-dev/mangahub_/cli/internal/update"
	"github.com/ngocan-dev/mangahub_/cli/internal/version"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:     "check",
	Short:   "Check for newer MangaHub CLI versions",
	Long:    "Contact the MangaHub update service to determine if a newer CLI release is available.",
	Example: "mangahub update check",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		info := updateclient.CheckForUpdates(version.CLIVersion)
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

		cmd.Println("Checking for updates...")
		cmd.Println()

		if info.UpToDate {
			cmd.Printf("You are running the latest version of MangaHub CLI (%s).\n", info.CurrentVersion)
			return nil
		}

		cmd.Printf("New version available: %s\n", info.LatestVersion)
		cmd.Printf("Current version:       %s\n\n", info.CurrentVersion)
		if len(info.ReleaseNotes) > 0 {
			cmd.Println("Changes:")
			for _, note := range info.ReleaseNotes {
				cmd.Printf("  - %s\n", note)
			}
			cmd.Println()
		}
		cmd.Println("To update:")
		cmd.Println("mangahub update install")
		return nil
	},
}

func init() {
	UpdateCmd.AddCommand(checkCmd)
	output.AddFlag(checkCmd)
}
