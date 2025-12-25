package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures CORS headers and handles preflight requests.
// Allowed origins are matched case-sensitively against the provided list.
// An origin of "*" allows all origins.
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowAll := false
	origins := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		o := strings.TrimSpace(origin)
		if o == "" {
			continue
		}
		if o == "*" {
			allowAll = true
		}
		origins[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && (allowAll || containsOrigin(origins, origin)) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
			c.Header("Access-Control-Max-Age", "600")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func containsOrigin(origins map[string]struct{}, origin string) bool {
	_, ok := origins[origin]
	return ok
}
