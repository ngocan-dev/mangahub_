package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	mu          sync.RWMutex
	clients     map[string]*clientLimiter
	rate        int           // requests per window
	window      time.Duration // time window
	cleanupTick *time.Ticker
}

type clientLimiter struct {
	tokens     int
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed
// window: time window for the rate limit
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientLimiter),
		rate:    rate,
		window:  window,
	}

	// Cleanup old entries every minute
	rl.cleanupTick = time.NewTicker(1 * time.Minute)
	go rl.cleanup()

	return rl
}

// cleanup removes old client limiters
func (rl *RateLimiter) cleanup() {
	for range rl.cleanupTick.C {
		rl.mu.Lock()
		now := time.Now()
		for key, limiter := range rl.clients {
			limiter.mu.Lock()
			if now.Sub(limiter.lastUpdate) > rl.window*2 {
				delete(rl.clients, key)
			}
			limiter.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// getClientLimiter gets or creates a limiter for a client
func (rl *RateLimiter) getClientLimiter(key string) *clientLimiter {
	rl.mu.RLock()
	limiter, exists := rl.clients[key]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check after acquiring write lock
		limiter, exists = rl.clients[key]
		if !exists {
			limiter = &clientLimiter{
				tokens:     rl.rate,
				lastUpdate: time.Now(),
			}
			rl.clients[key] = limiter
		}
		rl.mu.Unlock()
	}

	return limiter
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(key string) bool {
	limiter := rl.getClientLimiter(key)

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(limiter.lastUpdate)

	// Refill tokens based on elapsed time
	if elapsed >= rl.window {
		limiter.tokens = rl.rate
		limiter.lastUpdate = now
	} else {
		// Add tokens proportionally
		tokensToAdd := int(float64(rl.rate) * elapsed.Seconds() / rl.window.Seconds())
		if tokensToAdd > 0 {
			limiter.tokens = min(limiter.tokens+tokensToAdd, rl.rate)
			limiter.lastUpdate = now
		}
	}

	if limiter.tokens > 0 {
		limiter.tokens--
		return true
	}

	return false
}

// RateLimitMiddleware creates a Gin middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP address or user ID)
		clientKey := c.ClientIP()

		// If user is authenticated, use user ID instead
		if userID, exists := c.Get("user_id"); exists {
			clientKey = "user:" + string(rune(userID.(int64)))
		}

		if !rl.Allow(clientKey) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded. please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
