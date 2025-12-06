package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/chat"
	"github.com/spf13/cobra"
	"strings"
)

var sendCmd = &cobra.Command{
	Use:     "send",
	Short:   "Send a chat message",
	Long:    "Send a message to a MangaHub chat room.",
	Example: "mangahub chat send --manga-id one-piece --message 'Hello'",
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		message, _ := cmd.Flags().GetString("message")
		if strings.TrimSpace(message) == "" {
			return fmt.Errorf("message cannot be empty")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := chat.NewWSClient(mangaID)
		if err := client.Connect(ctx); err != nil {
			fmt.Println("✗ Unable to connect to chat server.")
			fmt.Println("Check if chat server is running: mangahub server status")
			return err
		}
		defer client.Close()

		if err := client.Send(chat.OutgoingMessage{Action: "message", Text: message, Room: mangaID}); err != nil {
			return err
		}

		roomLabel := "#general"
		if mangaID != "" {
			roomLabel = "#" + mangaID
		}
		cmd.Printf("Message sent to %s.\n", roomLabel)
		return nil
	},
}

func init() {
	ChatCmd.AddCommand(sendCmd)
	sendCmd.Flags().String("manga-id", "", "Manga-specific chat room (default: general chat)")
	sendCmd.Flags().String("message", "", "Message content")
}
