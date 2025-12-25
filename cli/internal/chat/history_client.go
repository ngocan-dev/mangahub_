package chat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"encoding/json"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
)

// HistoryClient retrieves chat history without opening a WebSocket connection.
type HistoryClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Verbose    bool
}

// NewHistoryClient returns a HistoryClient configured for the default chat API.
func NewHistoryClient(baseURL string, verbose bool) *HistoryClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = config.ResolveChatHTTPBase(config.DefaultConfig(""))
	}
	return &HistoryClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Verbose:    verbose,
	}
}

// Fetch retrieves history for the provided room and limit. An empty room
// indicates the general chat.
func (c *HistoryClient) Fetch(ctx context.Context, room string, limit int) (*HistoryResponse, []byte, error) {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}

	endpoint, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, nil, err
	}
	endpoint.Path = "/chat/history"

	q := endpoint.Query()
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if room != "" {
		q.Set("room", room)
	}
	endpoint.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, raw, err
	}
	if resp.StatusCode >= 400 {
		return nil, raw, errors.New(resp.Status)
	}

	var history HistoryResponse
	if err := json.Unmarshal(raw, &history); err != nil {
		return nil, raw, err
	}

	return &history, raw, nil
}
