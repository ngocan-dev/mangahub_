package progress

import "github.com/spf13/cobra"

var historyCmd = &cobra.Command{
	Use:     "history",
	Short:   "Show reading history",
	Long:    "Display historical reading activity for your library entries.",
	Example: "mangahub progress history --id 123",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement progress history
		cmd.Println("Progress history is not yet implemented.")
		return nil
	},
}

func init() {
	ProgressCmd.AddCommand(historyCmd)
	historyCmd.Flags().String("id", "", "Manga identifier")
}
