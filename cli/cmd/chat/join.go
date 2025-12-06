package chat

import "github.com/spf13/cobra"

// ChatCmd handles chat operations.
var ChatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Participate in MangaHub chat",
	Long:  "Join chat rooms, send messages, and view history.",
}

var joinCmd = &cobra.Command{
	Use:     "join",
	Short:   "Join a chat room",
	Long:    "Join a MangaHub chat room to participate in discussions.",
	Example: "mangahub chat join --room general",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement chat join
		cmd.Println("Chat join is not yet implemented.")
		return nil
	},
}

func init() {
	ChatCmd.AddCommand(joinCmd)
	joinCmd.Flags().String("room", "general", "Chat room to join")
}
