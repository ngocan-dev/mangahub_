package update

import (
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/ngocan-dev/mangahub_/cli/internal/output"
	updateclient "github.com/ngocan-dev/mangahub_/cli/internal/update"
	"github.com/ngocan-dev/mangahub_/cli/internal/version"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install the latest MangaHub CLI version",
	Long:    "Download and install the latest available MangaHub CLI release.",
	Example: "mangahub update install",
	RunE: func(cmd *cobra.Command, args []string) error {
		format, err := output.GetFormat(cmd)
		if err != nil {
			return err
		}

		result := updateclient.InstallLatest(version.CLIVersion)
		if format == output.FormatJSON {
			output.PrintJSON(cmd, result)
			return nil
		}

		if config.Runtime().Verbose {
			output.PrintJSON(cmd, result)
			return nil
		}

		if config.Runtime().Quiet {
			return nil
		}

		cmd.Println("Updating MangaHub CLI to latest version...")
		cmd.Println()
		for _, step := range result.Steps {
			cmd.Printf("✓ %s\n", step)
		}
		cmd.Println()
		cmd.Println("Update complete!")
		cmd.Println("Run: mangahub version")
		return nil
	},
}

func init() {
	UpdateCmd.AddCommand(installCmd)
	output.AddFlag(installCmd)
}
