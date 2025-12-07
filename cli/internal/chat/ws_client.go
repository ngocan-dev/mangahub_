package chat

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"
)

// WSClient manages a WebSocket connection to the chat server.
type WSClient struct {
	Room string

	conn   *simpleWS
	mu     sync.Mutex
	dialer *websocketDialer
}

type websocketDialer struct{}

// NewWSClient builds a WebSocket client for the desired room. An empty room
// value joins the general chat.
func NewWSClient(room string) *WSClient {
	return &WSClient{Room: room, dialer: &websocketDialer{}}
}

// Connect opens the WebSocket connection.
func (c *WSClient) Connect(ctx context.Context) error {
	values := url.Values{}
	if c.Room != "" {
		values.Set("room", c.Room)
	}

	endpoint := url.URL{Scheme: "ws", Host: "localhost:9093", Path: "/chat"}
	if encoded := values.Encode(); encoded != "" {
		endpoint.RawQuery = encoded
	}

	conn, err := c.dialer.DialContext(ctx, endpoint.String())
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	return nil
}

// Send transmits a message payload.
func (c *WSClient) Send(msg OutgoingMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("connection not established")
	}

	return c.conn.WriteJSON(msg)
}

// Read decodes the next incoming chat message.
func (c *WSClient) Read() (Message, error) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()
	if conn == nil {
		return Message{}, fmt.Errorf("connection not established")
	}

	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		return Message{}, err
	}
	return msg, nil
}

// Close cleanly shuts down the WebSocket connection.
func (c *WSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return nil
	}
	_ = c.conn.WriteControl(closeMessage, formatCloseMessage(closeNormalClosure, ""), time.Now().Add(time.Second))
	err := c.conn.Close()
	c.conn = nil
	return err
}

// Minimal WebSocket client implementation (text frames only).
type simpleWS struct {
	conn   net.Conn
	reader *bufio.Reader
}

type opcode byte

type controlMessage int

const (
	textMessage  opcode = 1
	closeMessage opcode = 8
	pingMessage  opcode = 9
	pongMessage  opcode = 10

	closeNormalClosure controlMessage = 1000
)

func formatCloseMessage(code controlMessage, text string) []byte {
	payload := make([]byte, 2+len(text))
	binary.BigEndian.PutUint16(payload, uint16(code))
	copy(payload[2:], text)
	return payload
}

func (d *websocketDialer) DialContext(ctx context.Context, endpoint string) (*simpleWS, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	host := u.Host
	if !strings.Contains(host, ":") {
		host += ":80"
	}

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return nil, err
	}

	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err != nil {
		conn.Close()
		return nil, err
	}
	secKey := base64.StdEncoding.EncodeToString(keyBytes)

	path := u.Path
	if path == "" {
		path = "/"
	}
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}

	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n", path, u.Host, secKey)
	if _, err := conn.Write([]byte(req)); err != nil {
		conn.Close()
		return nil, err
	}

	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return nil, err
	}
	if !strings.Contains(statusLine, "101") {
		conn.Close()
		return nil, fmt.Errorf("websocket handshake failed: %s", strings.TrimSpace(statusLine))
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			conn.Close()
			return nil, err
		}
		if line == "\r\n" {
			break
		}
	}

	return &simpleWS{conn: conn, reader: reader}, nil
}

func (c *simpleWS) writeFrame(op opcode, payload []byte) error {
	maskKey := make([]byte, 4)
	if _, err := rand.Read(maskKey); err != nil {
		return err
	}

	header := []byte{0x80 | byte(op)}
	payloadLen := len(payload)
	switch {
	case payloadLen < 126:
		header = append(header, byte(0x80|payloadLen))
	case payloadLen <= 65535:
		header = append(header, 0xFE)
		ext := make([]byte, 2)
		binary.BigEndian.PutUint16(ext, uint16(payloadLen))
		header = append(header, ext...)
	default:
		header = append(header, 0xFF)
		ext := make([]byte, 8)
		binary.BigEndian.PutUint64(ext, uint64(payloadLen))
		header = append(header, ext...)
	}
	header = append(header, maskKey...)

	masked := make([]byte, payloadLen)
	for i := range payload {
		masked[i] = payload[i] ^ maskKey[i%4]
	}

	err := connWrite(c.conn, append(header, masked...))
	return err
}

func connWrite(conn net.Conn, data []byte) error {
	total := 0
	for total < len(data) {
		n, err := conn.Write(data[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

func (c *simpleWS) WriteJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.writeFrame(textMessage, data)
}

func (c *simpleWS) ReadJSON(dest interface{}) error {
	for {
		op, payload, err := c.readFrame()
		if err != nil {
			return err
		}
		switch op {
		case textMessage:
			return json.Unmarshal(payload, dest)
		case pingMessage:
			_ = c.writeFrame(pongMessage, payload)
		case closeMessage:
			return io.EOF
		default:
			// ignore other opcodes
		}
	}
}

func (c *simpleWS) readFrame() (opcode, []byte, error) {
	b1, err := c.reader.ReadByte()
	if err != nil {
		return 0, nil, err
	}
	op := opcode(b1 & 0x0F)

	b2, err := c.reader.ReadByte()
	if err != nil {
		return 0, nil, err
	}
	masked := b2&0x80 != 0
	length := int(b2 & 0x7F)

	if length == 126 {
		var ext [2]byte
		if _, err := io.ReadFull(c.reader, ext[:]); err != nil {
			return 0, nil, err
		}
		length = int(binary.BigEndian.Uint16(ext[:]))
	} else if length == 127 {
		var ext [8]byte
		if _, err := io.ReadFull(c.reader, ext[:]); err != nil {
			return 0, nil, err
		}
		length64 := binary.BigEndian.Uint64(ext[:])
		if length64 > 1<<31 {
			return 0, nil, fmt.Errorf("frame too large")
		}
		length = int(length64)
	}

	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(c.reader, maskKey[:]); err != nil {
			return 0, nil, err
		}
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(c.reader, payload); err != nil {
		return 0, nil, err
	}
	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return op, payload, nil
}

func (c *simpleWS) WriteControl(op opcode, data []byte, _ time.Time) error {
	return c.writeFrame(op, data)
}

func (c *simpleWS) Close() error {
	return c.conn.Close()
}
