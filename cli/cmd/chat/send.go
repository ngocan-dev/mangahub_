package chat

import "github.com/spf13/cobra"

var sendCmd = &cobra.Command{
	Use:     "send",
	Short:   "Send a chat message",
	Long:    "Send a message to a MangaHub chat room.",
	Example: "mangahub chat send --room general --message 'Hello'",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement chat send
		cmd.Println("Chat send is not yet implemented.")
		return nil
	},
}

func init() {
	ChatCmd.AddCommand(sendCmd)
	sendCmd.Flags().String("room", "general", "Chat room name")
	sendCmd.Flags().String("message", "", "Message content")
}
