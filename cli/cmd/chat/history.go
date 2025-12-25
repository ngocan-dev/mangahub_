package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/chat"
	"github.com/ngocan-dev/mangahub_/cli/internal/config"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:     "history",
	Short:   "View chat history",
	Long:    "View recent chat messages from a MangaHub room.",
	Example: "mangahub chat history --manga-id one-piece",
	RunE: func(cmd *cobra.Command, args []string) error {
		mangaID, _ := cmd.Flags().GetString("manga-id")
		limit, _ := cmd.Flags().GetInt("limit")

		runtime := config.Runtime()
		cfg := config.ManagerInstance()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		client := chat.NewHistoryClient(config.ResolveChatHTTPBase(cfg.Data), runtime.Verbose)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		history, raw, err := client.Fetch(ctx, mangaID, limit)
		if runtime.Verbose && len(raw) > 0 {
			fmt.Println(string(raw))
		}
		if err != nil {
			fmt.Println("✗ Unable to load chat history.")
			fmt.Println("Check if chat server is running: mangahub server status")
			return err
		}

		chat.RenderHistory(history, mangaID, limit, runtime.Quiet)
		return nil
	},
}

func init() {
	ChatCmd.AddCommand(historyCmd)
	historyCmd.Flags().String("manga-id", "", "Manga-specific chat room (default: general chat)")
	historyCmd.Flags().Int("limit", 20, "Number of messages to show")
}
