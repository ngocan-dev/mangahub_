package library

import "github.com/spf13/cobra"

var batchAddCmd = &cobra.Command{
	Use:     "batch-add",
	Short:   "Add multiple manga entries",
	Long:    "Batch add manga entries to your library from a list of IDs or file.",
	Example: "mangahub library batch-add --file manga_ids.txt",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement batch add
		cmd.Println("Batch add is not yet implemented.")
		return nil
	},
}

func init() {
	LibraryCmd.AddCommand(batchAddCmd)
	batchAddCmd.Flags().String("file", "", "File containing manga IDs")
}
