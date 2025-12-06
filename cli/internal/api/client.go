package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a reusable HTTP API client for MangaHub.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient constructs a new API client with the provided base URL and token.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// Register registers a new user.
func (c *Client) Register(ctx context.Context, username, email, password string) (*RegisterResponse, error) {
	payload := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}
	var resp RegisterResponse
	err := c.doRequest(ctx, http.MethodPost, "/auth/register", payload, &resp)
	return &resp, err
}

// Login authenticates a user and returns the API token.
func (c *Client) Login(ctx context.Context, username string) (string, map[string]any, error) {
	payload := map[string]string{"username": username}
	var resp struct {
		Token string `json:"token"`
	}
	err := c.doRequest(ctx, http.MethodPost, "/auth/login", payload, &resp)
	return resp.Token, map[string]any{"token": resp.Token}, err
}

// Manga represents a manga search result.
type Manga struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// SearchManga searches manga titles.
func (c *Client) SearchManga(ctx context.Context, query string) ([]Manga, error) {
	endpoint := "/manga/search"
	u, _ := url.Parse(c.baseURL + endpoint)
	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	c.applyHeaders(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if err := checkStatus(res); err != nil {
		return nil, err
	}

	var results []Manga
	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		return nil, err
	}
	return results, nil
}

// AddToLibrary adds a manga to the library.
func (c *Client) AddToLibrary(ctx context.Context, mangaID, status string) (map[string]any, error) {
	payload := map[string]any{"manga_id": mangaID, "status": status}
	var resp map[string]any
	err := c.doRequest(ctx, http.MethodPost, "/library/add", payload, &resp)
	return resp, err
}

// UpdateProgress updates manga reading progress.
func (c *Client) UpdateProgress(ctx context.Context, mangaID string, chapter int) (map[string]any, error) {
	payload := map[string]any{"manga_id": mangaID, "chapter": chapter}
	var resp map[string]any
	err := c.doRequest(ctx, http.MethodPost, "/progress/update", payload, &resp)
	return resp, err
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, target any) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewBuffer(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	c.applyHeaders(req)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if err := checkStatus(res); err != nil {
		return err
	}

	if target != nil {
		if err := json.NewDecoder(res.Body).Decode(target); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) applyHeaders(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

func checkStatus(res *http.Response) error {
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	data, _ := io.ReadAll(res.Body)
	var apiErr struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(data, &apiErr)

	if apiErr.Error != "" {
		return &Error{Code: apiErr.Error, Message: apiErr.Message, Status: res.StatusCode}
	}

	return fmt.Errorf("api error: %s", strings.TrimSpace(string(data)))
}

// Error represents an error response from the MangaHub API.
type Error struct {
	Code    string
	Message string
	Status  int
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Code != "" {
		return fmt.Sprintf("api error: %s", e.Code)
	}
	return "api error"
}

// RegisterResponse represents the registration response payload.
type RegisterResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}
