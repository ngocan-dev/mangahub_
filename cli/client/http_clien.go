package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient quản lý các HTTP requests đến API
type HTTPClient struct {
	baseURL string
	token   string
	client  *http.Client
}

// LoginRequest
type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email"`
	Password        string `json:"password"`
}

// LoginResponse
type LoginResponse struct {
	Token string      `json:"token"`
	User  interface{} `json:"user"`
}

// MangaSearchResponse
type MangaSearchResponse struct {
	Results []Manga `json:"results"`
	Total   int     `json:"total"`
	Page    int     `json:"page"`
	Limit   int     `json:"limit"`
	Pages   int     `json:"pages"`
}

// Manga
type Manga struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Genre       string  `json:"genre"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	RatingPoint float64 `json:"rating_point"`
}

// MangaDetail
type MangaDetail struct {
	Manga
	ChapterCount  int         `json:"chapter_count"`
	Chapters      []Chapter   `json:"chapters,omitempty"`
	LibraryStatus interface{} `json:"library_status,omitempty"`
	UserProgress  interface{} `json:"user_progress,omitempty"`
}

// Chapter
type Chapter struct {
	ID         int64  `json:"id"`
	ChapterNum int    `json:"chapter_num"`
	Title      string `json:"title"`
	ReleasedAt string `json:"released_at"`
}

// NewHTTPClient
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken đặt authentication token
func (c *HTTPClient) SetToken(token string) {
	c.token = token
}

// Login đăng nhập và lấy token
func (c *HTTPClient) Login(usernameOrEmail, password string) (*LoginResponse, error) {
	reqBody := LoginRequest{
		UsernameOrEmail: usernameOrEmail,
		Password:        password,
	}

	var resp LoginResponse
	if err := c.post("/login", reqBody, &resp); err != nil {
		return nil, err
	}

	// Lưu token
	c.token = resp.Token
	return &resp, nil
}

// SearchManga tìm kiếm manga
func (c *HTTPClient) SearchManga(query string, page, limit int) (*MangaSearchResponse, error) {
	url := fmt.Sprintf("/manga/search?q=%s&page=%d&limit=%d", query, page, limit)

	var resp MangaSearchResponse
	if err := c.get(url, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetMangaDetails lấy chi tiết manga
func (c *HTTPClient) GetMangaDetails(mangaID int64) (*MangaDetail, error) {
	url := fmt.Sprintf("/manga/%d", mangaID)

	var resp MangaDetail
	if err := c.get(url, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetPopularManga lấy danh sách manga phổ biến
func (c *HTTPClient) GetPopularManga(limit int) (*MangaSearchResponse, error) {
	url := fmt.Sprintf("/manga/popular?limit=%d", limit)

	var resp MangaSearchResponse
	if err := c.get(url, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateProgress cập nhật tiến độ đọc
func (c *HTTPClient) UpdateProgress(mangaID int64, chapter int, chapterID *int64) error {
	url := fmt.Sprintf("/manga/%d/progress", mangaID)

	reqBody := map[string]interface{}{
		"chapter": chapter,
	}
	if chapterID != nil {
		reqBody["chapter_id"] = *chapterID
	}

	return c.put(url, reqBody, nil)
}

// AddToLibrary thêm manga vào thư viện
func (c *HTTPClient) AddToLibrary(mangaID int64) error {
	url := fmt.Sprintf("/manga/%d/library", mangaID)
	return c.post(url, nil, nil)
}

// get thực hiện HTTP GET request
func (c *HTTPClient) get(path string, result interface{}) error {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return err
	}

	return c.doRequest(req, result)
}

// post thực hiện HTTP POST request
func (c *HTTPClient) post(path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", c.baseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, result)
}

// put thực hiện HTTP PUT request
func (c *HTTPClient) put(path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("PUT", c.baseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req, result)
}

// doRequest thực hiện request và parse response
func (c *HTTPClient) doRequest(req *http.Request, result interface{}) error {
	// Thêm token nếu có
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response: %w", err)
	}

	// Kiểm tra status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse JSON
	if result != nil {
		if err := json.Unmarshal(bodyBytes, result); err != nil {
			return fmt.Errorf("cannot parse JSON: %w", err)
		}
	}

	return nil
}

// GetToken trả về token hiện tại
func (c *HTTPClient) GetToken() string {
	return c.token
}

// IsAuthenticated, check xác thực
func (c *HTTPClient) IsAuthenticated() bool {
	return c.token != ""
}
