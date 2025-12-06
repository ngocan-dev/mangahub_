package notify

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ngocan-dev/mangahub_/cli/internal/config"
)

// ErrTimeout is returned when waiting for a UDP message exceeds the deadline.
var ErrTimeout = errors.New("udp notification timeout")

// Preferences describe notification settings for the current device.
type Preferences struct {
	Enabled  bool
	Delivery string
	Events   []string
	Port     int
}

// UDPClient manages subscription and UDP helpers.
type UDPClient struct {
	port       int
	serverHost string
	cfg        *config.Manager
}

// NewUDPClient builds a UDP client using config defaults.
func NewUDPClient(cfg *config.Manager) *UDPClient {
	port := config.DefaultUDPPort
	if cfg != nil && cfg.Data.UDPPort > 0 {
		port = cfg.Data.UDPPort
	}
	return &UDPClient{port: port, serverHost: "127.0.0.1", cfg: cfg}
}

// Port returns the configured UDP port.
func (c *UDPClient) Port() int {
	return c.port
}

// Address returns the UDP address for the client.
func (c *UDPClient) Address() string {
	return fmt.Sprintf("%s:%d", c.serverHost, c.port)
}

// Subscribe enables local preferences and simulates backend registration.
func (c *UDPClient) Subscribe(ctx context.Context) (bool, error) {
	if c.cfg == nil {
		return false, errors.New("configuration not loaded")
	}

	if c.cfg.Data.Settings.Notifications {
		return false, nil
	}

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(150 * time.Millisecond):
	}

	c.cfg.Data.Settings.Notifications = true
	if err := c.cfg.Save(); err != nil {
		return false, err
	}
	return true, nil
}

// Unsubscribe disables local preferences and simulates backend unregistration.
func (c *UDPClient) Unsubscribe(ctx context.Context) (bool, error) {
	if c.cfg == nil {
		return false, errors.New("configuration not loaded")
	}

	if !c.cfg.Data.Settings.Notifications {
		return false, nil
	}

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(150 * time.Millisecond):
	}

	c.cfg.Data.Settings.Notifications = false
	if err := c.cfg.Save(); err != nil {
		return false, err
	}
	return true, nil
}

// Preferences returns a summary of the current notification configuration.
func (c *UDPClient) Preferences() Preferences {
	enabled := false
	if c.cfg != nil {
		enabled = c.cfg.Data.Settings.Notifications
	}
	return Preferences{
		Enabled:  enabled,
		Delivery: "UDP",
		Events:   []string{"New chapter releases", "Library updates"},
		Port:     c.port,
	}
}

// TriggerTestNotification simulates a backend-triggered UDP payload.
func (c *UDPClient) TriggerTestNotification(ctx context.Context, payload string) {
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(400 * time.Millisecond):
		}

		conn, err := net.Dial("udp", c.Address())
		if err != nil {
			return
		}
		defer conn.Close()

		_, _ = conn.Write([]byte(payload))
	}()
}

// WaitForTestNotification listens for a single UDP packet within the timeout.
func (c *UDPClient) WaitForTestNotification(ctx context.Context, timeout time.Duration) (string, error) {
	conn, err := net.ListenPacket("udp", fmt.Sprintf(":%d", c.port))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(timeout))

	go func() {
		<-ctx.Done()
		_ = conn.SetDeadline(time.Now())
	}()

	buf := make([]byte, 2048)
	n, _, readErr := conn.ReadFrom(buf)
	if readErr != nil {
		if strings.Contains(strings.ToLower(readErr.Error()), "timeout") {
			return "", ErrTimeout
		}
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", readErr
	}

	return string(buf[:n]), nil
}
