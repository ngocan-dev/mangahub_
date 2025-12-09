package chat

import "time"

// Message represents a chat message exchanged over WebSocket or returned by
// the history endpoint.
type Message struct {
	Timestamp time.Time `json:"timestamp"`
	User      string    `json:"user"`
	Text      string    `json:"text"`
	Room      string    `json:"room,omitempty"`
	Type      string    `json:"type,omitempty"`
	To        string    `json:"to,omitempty"`
}

// HistoryResponse models the payload returned by the chat history API.
type HistoryResponse struct {
	Room     string    `json:"room"`
	Messages []Message `json:"messages"`
}

// OutgoingMessage is sent to the WebSocket server to control chat actions.
type OutgoingMessage struct {
	Action  string `json:"action"`
	Text    string `json:"text,omitempty"`
	Room    string `json:"room,omitempty"`
	To      string `json:"to,omitempty"`
	MangaID string `json:"manga_id,omitempty"`
}
