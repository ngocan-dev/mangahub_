package chat

import (
	"fmt"
	"github.com/ngocan-dev/mangahub_/cli/internal/chat"
	"github.com/spf13/cobra"
)

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
	Example: "mangahub chat join --manga-id one-piece",
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		session := chat.NewSession(mangaID)
		if err := session.Run(); err != nil {
			fmt.Println("✗ Unable to connect to chat server.")
			fmt.Println("Check if chat server is running: mangahub server status")
			return err
		}
		return nil
	},
}

func init() {
	ChatCmd.AddCommand(joinCmd)
	joinCmd.Flags().String("manga-id", "", "Manga-specific chat room (default: general chat)")
}
