package client

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// UDPClient quản lý kết nối UDP với server
type UDPClient struct {
	serverAddr string
	conn       *net.UDPConn
	token      string
	userID     int64
	novelIDs   []int64
	allNovels  bool
	mu         sync.Mutex
	running    bool
	onNotify   func(ChapterNotification)
}

// ChapterNotification là thông báo chapter mới
type ChapterNotification struct {
	NovelID   int64  `json:"novel_id"`
	NovelName string `json:"novel_name"`
	Chapter   int    `json:"chapter"`
	ChapterID int64  `json:"chapter_id"`
	Timestamp string `json:"timestamp"`
}

// UDPPacket là cấu trúc packet UDP
type UDPPacket struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewUDPClient tạo UDP client mới
func NewUDPClient(serverAddr, token string, userID int64) *UDPClient {
	return &UDPClient{
		serverAddr: serverAddr,
		token:      token,
		userID:     userID,
		allNovels:  true, // Mặc định nhận tất cả thông báo
	}
}

// SubscribeToNovels đăng ký nhận thông báo cho các manga cụ thể
func (c *UDPClient) SubscribeToNovels(novelIDs []int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.novelIDs = novelIDs
	c.allNovels = false
}

// SubscribeToAll đăng ký nhận tất cả thông báo
func (c *UDPClient) SubscribeToAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.allNovels = true
	c.novelIDs = nil
}

// Connect kết nối và đăng ký với UDP server
func (c *UDPClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Resolve địa chỉ server
	serverUDPAddr, err := net.ResolveUDPAddr("udp", c.serverAddr)
	if err != nil {
		return fmt.Errorf("không thể resolve địa chỉ: %w", err)
	}

	// Tạo kết nối UDP
	conn, err := net.DialUDP("udp", nil, serverUDPAddr)
	if err != nil {
		return fmt.Errorf("không thể kết nối UDP: %w", err)
	}
	c.conn = conn
	c.running = true

	// Gửi packet đăng ký
	registerPacket := UDPPacket{
		Type: "register",
		Payload: map[string]interface{}{
			"user_id":    c.userID,
			"token":      c.token,
			"novel_ids":  c.novelIDs,
			"all_novels": c.allNovels,
			"device_id":  "cli-client",
		},
	}

	if err := c.sendPacket(registerPacket); err != nil {
		conn.Close()
		return fmt.Errorf("không thể gửi đăng ký: %w", err)
	}

	// Đợi xác nhận
	packet, err := c.receivePacket(5 * time.Second)
	if err != nil {
		conn.Close()
		return fmt.Errorf("không nhận được xác nhận: %w", err)
	}

	if packet.Type != "confirm" {
		conn.Close()
		return fmt.Errorf("đăng ký thất bại: %s", packet.Error)
	}

	fmt.Println("✓ Đã đăng ký nhận thông báo UDP")

	// Bắt đầu nhận thông báo
	go c.receiveNotifications()

	return nil
}

// SetNotificationCallback đặt hàm callback khi nhận thông báo
func (c *UDPClient) SetNotificationCallback(callback func(ChapterNotification)) {
	c.onNotify = callback
}

// receiveNotifications nhận và xử lý thông báo từ server
func (c *UDPClient) receiveNotifications() {
	buffer := make([]byte, 4096)

	for c.running {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		n, err := c.conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Timeout - tiếp tục chờ
				continue
			}
			if c.running {
				fmt.Printf("⚠ Lỗi đọc UDP: %v\n", err)
			}
			break
		}

		// Parse packet
		var packet UDPPacket
		if err := json.Unmarshal(buffer[:n], &packet); err != nil {
			continue
		}

		// Xử lý thông báo
		if packet.Type == "notification" {
			c.handleNotification(packet)
		} else if packet.Type == "error" {
			fmt.Printf("⚠ Lỗi từ server: %s\n", packet.Error)
		}
	}
}

// handleNotification xử lý thông báo chapter mới
func (c *UDPClient) handleNotification(packet UDPPacket) {
	// Chuyển payload thành ChapterNotification
	payloadBytes, err := json.Marshal(packet.Payload)
	if err != nil {
		return
	}

	var notif ChapterNotification
	if err := json.Unmarshal(payloadBytes, &notif); err != nil {
		return
	}

	// Gọi callback nếu có
	if c.onNotify != nil {
		c.onNotify(notif)
	}
}

// sendPacket gửi packet đến server
func (c *UDPClient) sendPacket(packet UDPPacket) error {
	data, err := json.Marshal(packet)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(data)
	return err
}

// receivePacket nhận một packet với timeout
func (c *UDPClient) receivePacket(timeout time.Duration) (*UDPPacket, error) {
	c.conn.SetReadDeadline(time.Now().Add(timeout))

	buffer := make([]byte, 4096)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	var packet UDPPacket
	if err := json.Unmarshal(buffer[:n], &packet); err != nil {
		return nil, err
	}

	return &packet, nil
}

// Unregister hủy đăng ký nhận thông báo
func (c *UDPClient) Unregister() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running || c.conn == nil {
		return nil
	}

	// Gửi packet hủy đăng ký
	unregisterPacket := UDPPacket{
		Type: "unregister",
		Payload: map[string]interface{}{
			"user_id": c.userID,
			"token":   c.token,
		},
	}

	return c.sendPacket(unregisterPacket)
}

// Close đóng kết nối
func (c *UDPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.running = false
	if c.conn != nil {
		// Unregister trước khi đóng
		unregisterPacket := UDPPacket{
			Type: "unregister",
			Payload: map[string]interface{}{
				"user_id": c.userID,
			},
		}
		c.sendPacket(unregisterPacket)

		return c.conn.Close()
	}
	return nil
}

// IsConnected kiểm tra kết nối
func (c *UDPClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.running && c.conn != nil
}
