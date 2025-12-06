package update

import "github.com/spf13/cobra"

var installCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install updates",
	Long:    "Download and install available MangaHub CLI updates.",
	Example: "mangahub update install",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement update install
		cmd.Println("Update install is not yet implemented.")
		return nil
	},
}

func init() {
	UpdateCmd.AddCommand(installCmd)
	installCmd.Flags().Bool("pre", false, "Include pre-release versions")
}
