package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// TCPClient quản lý kết nối TCP với server
type TCPClient struct {
	address    string
	conn       net.Conn
	token      string
	deviceName string
	deviceType string
	mu         sync.Mutex
	running    bool
	onProgress func(ProgressUpdate)
}

// ProgressUpdate là thông tin cập nhật tiến độ đọc
type ProgressUpdate struct {
	UserID    int64  `json:"user_id"`
	NovelID   int64  `json:"novel_id"`
	Chapter   int    `json:"chapter"`
	ChapterID *int64 `json:"chapter_id,omitempty"`
	Timestamp string `json:"timestamp"`
}

// TCPMessage là cấu trúc message TCP
type TCPMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewTCPClient tạo TCP client mới
func NewTCPClient(address, token string) *TCPClient {
	return &TCPClient{
		address:    address,
		token:      token,
		deviceName: "CLI",
		deviceType: "terminal",
	}
}

// Connect kết nối đến TCP server
func (c *TCPClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Kết nối TCP
	conn, err := net.DialTimeout("tcp", c.address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("Can not connect: %w", err)
	}
	c.conn = conn
	c.running = true

	// Gửi message xác thực
	authMsg := TCPMessage{
		Type: "auth",
		Payload: map[string]string{
			"token":       c.token,
			"device_name": c.deviceName,
			"device_type": c.deviceType,
		},
	}

	// nhận xác thực
	if err := c.sendMessage(authMsg); err != nil {
		conn.Close()
		return fmt.Errorf("cannot send auth: %w", err)
	}

	// Đọc phản hồi xác thực
	response, err := c.readMessage()
	if err != nil {
		conn.Close()
		return fmt.Errorf("cannot receive auth response: %w", err)
	}

	if response.Type != "auth_response" {
		conn.Close()
		return fmt.Errorf("auth failed: %s", response.Error)
	}

	fmt.Println("✓ Connected and authenticated TCP")

	// Bắt đầu nhận messages
	go c.receiveMessages()

	return nil
}

// SetProgressCallback đặt hàm callback khi nhận progress update
func (c *TCPClient) SetProgressCallback(callback func(ProgressUpdate)) {
	c.onProgress = callback
}

// receiveMessages nhận và xử lý messages từ server
func (c *TCPClient) receiveMessages() {
	scanner := bufio.NewScanner(c.conn)
	for c.running && scanner.Scan() {
		line := scanner.Bytes()

		var msg TCPMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}

		// Xử lý các loại message
		switch msg.Type {
		case "progress":
			c.handleProgressUpdate(msg)
		case "heartbeat":
			// Server gửi heartbeat, không cần làm gì
		case "error":
			fmt.Printf("⚠ Lỗi từ server: %s\n", msg.Error)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("⚠ Lỗi đọc dữ liệu: %v\n", err)
	}
}

// handleProgressUpdate xử lý progress update
func (c *TCPClient) handleProgressUpdate(msg TCPMessage) {
	// Chuyển payload thành ProgressUpdate
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		return
	}

	var progress ProgressUpdate
	if err := json.Unmarshal(payloadBytes, &progress); err != nil {
		return
	}

	// Gọi callback nếu có
	if c.onProgress != nil {
		c.onProgress(progress)
	}
}

// sendMessage gửi message đến server
func (c *TCPClient) sendMessage(msg TCPMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Thêm newline delimiter
	data = append(data, '\n')

	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = c.conn.Write(data)
	return err
}

// readMessage đọc một message từ server
func (c *TCPClient) readMessage() (*TCPMessage, error) {
	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	scanner := bufio.NewScanner(c.conn)
	if scanner.Scan() {
		var msg TCPMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			return nil, err
		}
		return &msg, nil
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("không có dữ liệu")
}

// Close đóng kết nối
func (c *TCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.running = false
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected kiểm tra kết nối
func (c *TCPClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running && c.conn != nil
}
