package client

import "net/http"

// HTTPClient is a placeholder for MangaHub HTTP operations.
type HTTPClient struct {
	BaseURL string
	Client  *http.Client
}

// NewHTTPClient creates a new HTTP client.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

// Ping checks connectivity to the server.
func (c *HTTPClient) Ping() error {
	// TODO: implement HTTP ping logic
	return nil
}
