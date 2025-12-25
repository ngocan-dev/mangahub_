package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ngocan-dev/mangahub/backend/domain/user"
	"github.com/ngocan-dev/mangahub/backend/internal/auth"
)

type AuthHandler struct {
	DB *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

// Login handles user login
// Main Success Scenario:
// 1. User provides email and password
// 2. System validates credentials against database
// 3. System generates JWT token with user information
// 4. System returns token for subsequent requests
// 5. User can access protected endpoints
func (h *AuthHandler) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("login: invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email and password are required",
		})
		return
	}

	// Call login service
	response, err := user.Login(c.Request.Context(), h.DB, req)
	if err != nil {
		// A1: Invalid credentials
		if errors.Is(err, user.ErrInvalidCredentials) {
			log.Printf("login: invalid credentials for email=%s", req.Email)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid credentials",
			})
			return
		}

		// A2: Account not found - System suggests registration
		if errors.Is(err, user.ErrUserNotFound) {
			log.Printf("login: user not found for email=%s", req.Email)
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "account not found",
				"message": "please register to create an account",
			})
			return
		}

		// Other errors
		log.Printf("login: internal error for email=%s: %v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// Success: Return token and user info
	c.JSON(http.StatusOK, response)
}

// RequireAuth is a middleware that validates JWT token and sets user context
// Ensure only authenticated users access protected resources
// Invalid tokens are rejected
// Expired tokens trigger reauthentication
// Token claims are properly validated
// Unauthorized access is prevented
func (h *AuthHandler) RequireAuth(c *gin.Context) {
	// Get token from header or websocket-friendly locations
	tokenString := getTokenFromRequest(c)
	if tokenString == "" {
		// Unauthorized access is prevented
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "authorization token required",
			"message": "please provide a valid authentication token",
		})
		c.Abort()
		return
	}

	// Validate token
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		// Handle different error types with appropriate messages
		// Invalid tokens are rejected
		// Expired tokens trigger reauthentication
		if errors.Is(err, auth.ErrExpiredToken) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "expired token",
				"message": "your session has expired. please login again",
				"code":    "TOKEN_EXPIRED",
			})
			c.Abort()
			return
		}
		if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrInvalidSigningMethod) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid token",
				"message": "the provided token is invalid or malformed",
				"code":    "TOKEN_INVALID",
			})
			c.Abort()
			return
		}
		if errors.Is(err, auth.ErrInvalidClaims) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid token claims",
				"message": "the token contains invalid or missing claims",
				"code":    "TOKEN_CLAIMS_INVALID",
			})
			c.Abort()
			return
		}
		if errors.Is(err, auth.ErrTokenNotBefore) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "token not yet valid",
				"message": "the token is not yet valid",
				"code":    "TOKEN_NOT_BEFORE",
			})
			c.Abort()
			return
		}
		// Generic error
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "authentication failed",
			"message": "unable to validate authentication token",
		})
		c.Abort()
		return
	}

	if claims == nil {
		// Unauthorized access is prevented
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid token",
			"message": "token validation failed",
		})
		c.Abort()
		return
	}

	// Additional validation: ensure claims are not empty
	// Token claims are properly validated
	if claims.UserID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid token claims",
			"message": "token contains invalid user information",
		})
		c.Abort()
		return
	}

	// Set user context for downstream handlers
	c.Set("user_id", claims.UserID)
	c.Set("userID", claims.UserID) // compatibility with existing handlers
	c.Set("username", claims.Username)
	c.Set("email", claims.Email)

	c.Next()
}

// Me returns authenticated user info based on the JWT claims.
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userIDInt, ok := userID.(int64)
	if !ok || userIDInt <= 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var u struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	err := h.DB.QueryRowContext(ctx, `
		SELECT id, username, email
		FROM users
		WHERE id = ?
	`, userIDInt).Scan(&u.ID, &u.Username, &u.Email)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		log.Printf("me: failed to load user %d: %v", userIDInt, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, u)
}

func getTokenFromRequest(c *gin.Context) string {
	// Prefer standard Authorization header
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader != "" {
		lowered := strings.ToLower(authHeader)
		if strings.HasPrefix(lowered, "bearer ") {
			return strings.TrimSpace(authHeader[len("Bearer "):])
		}
	}

	// WebSocket clients cannot set custom headers; allow token via query/subprotocol
	if strings.EqualFold(c.GetHeader("Upgrade"), "websocket") || strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade") {
		if token := strings.TrimSpace(c.Query("token")); token != "" {
			return token
		}

		if proto := c.GetHeader("Sec-WebSocket-Protocol"); proto != "" {
			parts := strings.Split(proto, ",")
			if len(parts) > 0 {
				return strings.TrimSpace(parts[0])
			}
		}
	}

	return ""
}
