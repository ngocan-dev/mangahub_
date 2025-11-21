package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ngocan-dev/mangahub/backend/internal/auth"
	"github.com/ngocan-dev/mangahub/backend/internal/security"
)

var (
	emailRegex            = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	ErrDuplicateUsername  = errors.New("duplicate username")
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

// A2: invalid email
// Invalid data formats are rejected
func ValidateEmail(email string) error {
	return security.ValidateEmail(email)
}

// A3: weak password
// Invalid data formats are rejected
// Input length limits are enforced
func ValidatePassword(password string) error {
	return security.ValidatePassword(password)
}

// Main success path + A1
func Register(ctx context.Context, db *sql.DB, req RegistrationRequest) (*User, error) {
	// Validate username format
	// Invalid data formats are rejected
	// SQL injection attempts are blocked
	if err := security.ValidateUsername(req.Username); err != nil {
		return nil, fmt.Errorf("invalid username: %w", err)
	}

	// Validate email format
	// Invalid data formats are rejected
	if err := security.ValidateEmail(req.Email); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Validate password strength
	if err := security.ValidatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// A1: check username trùng
	// SQL injection attempts are blocked (parameterized query)
	var existing int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(1) FROM Users WHERE Username = ?",
		req.Username,
	).Scan(&existing); err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, ErrDuplicateUsername
	}

	// Check duplicate email
	existing = 0
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(1) FROM Users WHERE Email = ?",
		req.Email,
	).Scan(&existing); err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, ErrDuplicateEmail
	}

	// Step 3: hash mật khẩu
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now().UTC()

	// Step 4: insert DB
	res, err := db.ExecContext(ctx, `
                INSERT INTO Users (Username, Email, PasswordHash, Created_Date)
                VALUES (?, ?, ?, ?)
        `, req.Username, req.Email, string(hashed), createdAt)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Username: req.Username,
		Email:    req.Email,
	}, nil
}

// Login authenticates a user and returns user info
// Main Success Scenario:
// 1. User provides username/email and password
// 2. System validates credentials against database
// 3. System generates JWT token with user information
// 4. System returns token for subsequent requests
func Login(ctx context.Context, db *sql.DB, req LoginRequest) (*LoginResponse, error) {
	var userID int64
	var username, email, passwordHash string

	// Step 1 & 2: Find user by username or email and get password hash
	// Check if input is email or username
	isEmail := emailRegex.MatchString(req.UsernameOrEmail)

	var query string
	if isEmail {
		query = `SELECT UserId, Username, Email, PasswordHash FROM Users WHERE Email = ?`
	} else {
		query = `SELECT UserId, Username, Email, PasswordHash FROM Users WHERE Username = ?`
	}

	err := db.QueryRowContext(ctx, query, req.UsernameOrEmail).Scan(
		&userID, &username, &email, &passwordHash,
	)

	// A2: Account not found
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	// Step 2: Validate password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	// A1: Invalid credentials
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Step 3: Generate JWT token
	// Import auth package for token generation
	// We'll need to import it at the top
	token, err := generateTokenForUser(userID, username, email)
	if err != nil {
		return nil, err
	}

	// Step 4: Return token and user info
	return &LoginResponse{
		Token: token,
		User: &User{
			ID:       userID,
			Username: username,
			Email:    email,
		},
	}, nil
}

// generateTokenForUser is a helper to generate token
func generateTokenForUser(userID int64, username, email string) (string, error) {
	return auth.GenerateToken(userID, username, email)
}
