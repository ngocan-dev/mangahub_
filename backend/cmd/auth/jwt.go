package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrExpiredToken         = errors.New("expired token")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrInvalidClaims        = errors.New("invalid token claims")
	ErrTokenNotBefore       = errors.New("token not yet valid")
	jwtSecret               = []byte(getJWTSecret())
)

// getJWTSecret retrieves JWT secret from environment or uses default
func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Default secret for development - should be changed in production
		secret = "mangahub-secret-key-change-in-production"
	}
	return secret
}

// Claims represents JWT claims
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token for a user
func GenerateToken(userID int64, username, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mangahub",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
// Invalid tokens are rejected
// Expired tokens trigger reauthentication
// Token claims are properly validated
func ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method - only accept HS256
		// Invalid tokens are rejected
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return jwtSecret, nil
	})

	// Handle parsing errors
	if err != nil {
		// Check for specific JWT validation errors
		if errors.Is(err, jwt.ErrTokenExpired) {
			// Expired tokens trigger reauthentication
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotBefore
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrInvalidToken
		}
		// Other parsing errors
		return nil, ErrInvalidToken
	}

	// Validate token is valid
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate claims are present and valid
	// Token claims are properly validated
	if claims.UserID == 0 {
		return nil, ErrInvalidClaims
	}
	if claims.Username == "" {
		return nil, ErrInvalidClaims
	}
	if claims.Email == "" {
		return nil, ErrInvalidClaims
	}

	// Validate issuer
	if claims.Issuer != "mangahub" {
		return nil, ErrInvalidClaims
	}

	// Validate subject matches username
	if claims.Subject != claims.Username {
		return nil, ErrInvalidClaims
	}

	// Validate expiration time is set
	if claims.ExpiresAt == nil {
		return nil, ErrInvalidClaims
	}

	// Double-check expiration (jwt library should handle this, but we verify)
	if claims.ExpiresAt.Time.Before(time.Now()) {
		// Expired tokens trigger reauthentication
		return nil, ErrExpiredToken
	}

	// Validate not-before time
	if claims.NotBefore != nil && claims.NotBefore.Time.After(time.Now()) {
		return nil, ErrTokenNotBefore
	}

	return claims, nil
}
