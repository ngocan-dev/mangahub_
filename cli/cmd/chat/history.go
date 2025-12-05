package chat

import "github.com/spf13/cobra"

var historyCmd = &cobra.Command{
	Use:     "history",
	Short:   "View chat history",
	Long:    "View recent chat messages from a MangaHub room.",
	Example: "mangahub chat history --room general",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement chat history
		cmd.Println("Chat history is not yet implemented.")
		return nil
	},
}

func init() {
	ChatCmd.AddCommand(historyCmd)
	historyCmd.Flags().String("room", "general", "Chat room to view")
	historyCmd.Flags().Int("limit", 50, "Number of messages to show")
}
