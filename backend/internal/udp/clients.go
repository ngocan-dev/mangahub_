package udp

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Client represents a registered UDP client
type Client struct {
	Address      *net.UDPAddr
	UserID       int64
	NovelIDs     []int64 // Empty means all novels
	AllNovels    bool
	DeviceID     string
	RegisteredAt time.Time
	LastSeen     time.Time
	mu           sync.RWMutex
}

// NewClient creates a new UDP client instance
func NewClient(addr *net.UDPAddr, userID int64, novelIDs []int64, allNovels bool, deviceID string) *Client {
	now := time.Now()
	return &Client{
		Address:      addr,
		UserID:       userID,
		NovelIDs:     novelIDs,
		AllNovels:    allNovels,
		DeviceID:     deviceID,
		RegisteredAt: now,
		LastSeen:     now,
	}
}

// UpdateLastSeen updates the last seen timestamp
func (c *Client) UpdateLastSeen() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastSeen = time.Now()
}

// IsSubscribedTo checks if client is subscribed to a specific novel
func (c *Client) IsSubscribedTo(novelID int64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.AllNovels {
		return true
	}

	for _, id := range c.NovelIDs {
		if id == novelID {
			return true
		}
	}
	return false
}

// GetKey returns a unique key for the client (address + userID)
func (c *Client) GetKey() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Address.String() + ":" + fmt.Sprintf("%d", c.UserID)
}
