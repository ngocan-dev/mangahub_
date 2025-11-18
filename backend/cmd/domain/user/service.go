package user

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	emailRegex           = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	ErrDuplicateUsername = errors.New("duplicate username")
)

// A2: invalid email
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}

// A3: weak password
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return errors.New("password must include upper, lower, digit, and special characters")
	}
	return nil
}

// Main success path + A1
func Register(ctx context.Context, db *sql.DB, req RegistrationRequest) (*User, error) {
	// A1: check username trùng
	var existing int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(1) FROM users WHERE username = ?",
		req.Username,
	).Scan(&existing); err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, ErrDuplicateUsername
	}

	// Step 3: hash mật khẩu
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now().UTC()

	// Step 4: insert DB
	res, err := db.ExecContext(ctx, `
		INSERT INTO users (username, email, password_hash, created_at)
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
