package client

import (
	"fmt"
	"net/http"
	"time"
)

// HTTPClient is a placeholder for MangaHub HTTP operations.
type HTTPClient struct {
	BaseURL string
	Client  *http.Client
}

// NewHTTPClient creates a new HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Ping checks connectivity to the server.
func (c *HTTPClient) Ping() error {
	url := c.BaseURL + "/health"
	resp, err := c.Client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to ping server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// SetTimeout sets the HTTP client timeout.
func (c *HTTPClient) SetTimeout(timeout time.Duration) {
	c.Client.Timeout = timeout
}
