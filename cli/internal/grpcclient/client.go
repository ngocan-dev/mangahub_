package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
)

// Client wraps the reusable gRPC connection and service clients.
type Client struct {
	address  string
	manga    MangaServiceClient
	progress ProgressServiceClient
}

// ConnectionError represents a failure to reach the gRPC server.
type ConnectionError struct {
	Address string
	Err     error
}

// Error implements the error interface.
func (e *ConnectionError) Error() string {
	return fmt.Sprintf("could not connect to server at %s: %v", e.Address, e.Err)
}

// Unwrap exposes the wrapped error.
func (e *ConnectionError) Unwrap() error {
	return e.Err
}

// New creates a gRPC client using configuration defaults.
func New(ctx context.Context, cfg *config.Manager) (*Client, error) {
	address := config.DefaultGRPCAddress
	if cfg != nil && strings.TrimSpace(cfg.Data.GRPCAddress) != "" {
		address = cfg.Data.GRPCAddress
	}

	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := probeAddress(dialCtx, address); err != nil {
		return nil, err
	}

	return &Client{
		address:  address,
		manga:    newMockMangaClient(),
		progress: newMockProgressClient(),
	}, nil
}

// Close closes the underlying gRPC connection.
func (c *Client) Close() error {
	return nil
}

// Address returns the configured gRPC server address.
func (c *Client) Address() string {
	return c.address
}

// Manga exposes the MangaService client.
func (c *Client) Manga() MangaServiceClient {
	return c.manga
}

// Progress exposes the ProgressService client.
func (c *Client) Progress() ProgressServiceClient {
	return c.progress
}

func probeAddress(ctx context.Context, address string) error {
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return &ConnectionError{Address: address, Err: err}
	}
	_ = conn.Close()
	return nil
}

// Manga represents a manga entity returned by the gRPC API.
type Manga struct {
	Id            string
	Title         string
	OriginalTitle string
	Author        string
	Status        string
	Chapters      string
	Genres        []string
}

// GetMangaRequest wraps the identifier for manga lookups.
type GetMangaRequest struct {
	Id string
}

// SearchMangaRequest contains parameters for manga search.
type SearchMangaRequest struct {
	Query string
}

// SearchResult represents a single manga search result.
type SearchResult struct {
	Id       string
	Title    string
	Status   string
	Chapters string
}

// SearchMangaResponse wraps a collection of search results.
type SearchMangaResponse struct {
	Results []SearchResult
}

// UpdateProgressRequest is used to update reading progress.
type UpdateProgressRequest struct {
	MangaID string
	Chapter int
	Token   string
}

// UpdateProgressResponse reports the updated progress.
type UpdateProgressResponse struct {
	MangaTitle string
	Chapter    int
}

// MangaServiceClient is a minimal interface for manga RPCs.
type MangaServiceClient interface {
	GetManga(ctx context.Context, in *GetMangaRequest) (*Manga, error)
	SearchManga(ctx context.Context, in *SearchMangaRequest) (*SearchMangaResponse, error)
}

// ProgressServiceClient is a minimal interface for progress RPCs.
type ProgressServiceClient interface {
	UpdateProgress(ctx context.Context, in *UpdateProgressRequest) (*UpdateProgressResponse, error)
}

type mockMangaClient struct{}

type mockProgressClient struct{}

func newMockMangaClient() MangaServiceClient {
	return &mockMangaClient{}
}

func newMockProgressClient() ProgressServiceClient {
	return &mockProgressClient{}
}

func (m *mockMangaClient) GetManga(ctx context.Context, in *GetMangaRequest) (*Manga, error) {
	dataset := map[string]Manga{
		"one-piece": {
			Id:            "one-piece",
			Title:         "One Piece",
			OriginalTitle: "ワンピース",
			Author:        "Oda Eiichiro",
			Status:        "Ongoing",
			Chapters:      "1100+",
			Genres:        []string{"Action", "Adventure", "Comedy", "Shounen"},
		},
		"attack-on-titan": {
			Id:            "attack-on-titan",
			Title:         "Attack on Titan",
			OriginalTitle: "進撃の巨人",
			Author:        "Hajime Isayama",
			Status:        "Completed",
			Chapters:      "139",
			Genres:        []string{"Action", "Drama", "Fantasy"},
		},
	}

	if manga, ok := dataset[in.Id]; ok {
		return &manga, nil
	}

	return nil, fmt.Errorf("Manga not found: '%s'", in.Id)
}

func (m *mockMangaClient) SearchManga(ctx context.Context, in *SearchMangaRequest) (*SearchMangaResponse, error) {
	query := strings.ToLower(strings.TrimSpace(in.Query))
	if query == "" {
		return &SearchMangaResponse{Results: nil}, nil
	}

	if strings.Contains(query, "attack on titan") {
		return &SearchMangaResponse{Results: []SearchResult{
			{Id: "attack-on-titan", Title: "Attack on Titan", Status: "Completed", Chapters: "139"},
			{Id: "attack-on-titan-jr", Title: "Attack on Titan: Junior High", Status: "Completed", Chapters: "7"},
			{Id: "aot-before-fall", Title: "Attack on Titan: Before the Fall", Status: "Completed", Chapters: "17"},
		}}, nil
	}

	return &SearchMangaResponse{Results: nil}, nil
}

func (m *mockProgressClient) UpdateProgress(ctx context.Context, in *UpdateProgressRequest) (*UpdateProgressResponse, error) {
	if in == nil || strings.TrimSpace(in.MangaID) == "" {
		return nil, errors.New("manga id is required")
	}
	if in.Chapter <= 0 {
		return nil, errors.New("chapter must be greater than 0")
	}

	title := strings.ReplaceAll(in.MangaID, "-", " ")
	title = strings.Title(title)

	return &UpdateProgressResponse{
		MangaTitle: title,
		Chapter:    in.Chapter,
	}, nil
}
