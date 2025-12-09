package chat

import (
	"fmt"
	"time"
)

func RenderHistory(history *HistoryResponse, room string, limit int, quiet bool) {
	roomName := "general"
	if room != "" {
		roomName = room
	} else if history != nil && history.Room != "" {
		roomName = history.Room
	}

	if history == nil || len(history.Messages) == 0 {
		fmt.Printf("No chat history found for #%s.\n", roomName)
		return
	}

	if !quiet {
		fmt.Printf("Chat History: #%s (%d messages)\n\n", roomName, limit)
	}

	for _, msg := range history.Messages {
		prefix := msg.Timestamp.Local().Format("15:04")
		fmt.Printf("[%s] %s: %s\n", prefix, msg.User, msg.Text)
	}

	if quiet {
		return
	}

	fmt.Println("─────────────────────────────────────────────────────────────")
	if roomName == "general" {
		fmt.Println("To join live chat:")
		fmt.Println("mangahub chat join")
	} else {
		fmt.Println("To join this discussion live:")
		fmt.Printf("mangahub chat join --manga-id %s\n", roomName)
	}
}

// ParseTimestamp converts RFC3339 timestamps to the desired HH:MM presentation.
func ParseTimestamp(raw string) time.Time {
	if raw == "" {
		return time.Now()
	}
	ts, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Now()
	}
	return ts
}
