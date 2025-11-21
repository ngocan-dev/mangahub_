package user

import (
	"context"
	"database/sql"
	"testing"

	"golang.org/x/crypto/bcrypt"

	_ "modernc.org/sqlite"

	"github.com/ngocan-dev/mangahub/backend/internal/auth"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", "file:test_login?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	schema := `
CREATE TABLE Users (
    UserId INTEGER PRIMARY KEY AUTOINCREMENT,
    Username TEXT NOT NULL UNIQUE,
    PasswordHash TEXT NOT NULL,
    Email TEXT NOT NULL UNIQUE,
    Created_Date DATETIME DEFAULT CURRENT_TIMESTAMP
);`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func createTestUser(t *testing.T, db *sql.DB, username, email, password string) int64 {
	t.Helper()

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	res, err := db.Exec(`INSERT INTO Users (Username, Email, PasswordHash) VALUES (?, ?, ?)`, username, email, string(hashed))
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("failed to read user id: %v", err)
	}

	return id
}

func TestLoginSuccess(t *testing.T) {
	db := setupTestDB(t)
	userID := createTestUser(t, db, "reader1", "reader1@example.com", "password123")

	resp, err := Login(context.Background(), db, LoginRequest{UsernameOrEmail: "reader1", Password: "password123"})
	if err != nil {
		t.Fatalf("expected login to succeed, got error: %v", err)
	}

	if resp.User == nil {
		t.Fatalf("expected user info in response")
	}

	if resp.User.ID != userID || resp.User.Username != "reader1" || resp.User.Email != "reader1@example.com" {
		t.Fatalf("unexpected user data in response: %+v", resp.User)
	}

	if resp.Token == "" {
		t.Fatalf("expected token in response")
	}

	claims, err := auth.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("expected token to be valid, got error: %v", err)
	}

	if claims.UserID != userID || claims.Username != "reader1" || claims.Email != "reader1@example.com" {
		t.Fatalf("unexpected claims in token: %+v", claims)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	db := setupTestDB(t)
	createTestUser(t, db, "reader2", "reader2@example.com", "correct-password")

	_, err := Login(context.Background(), db, LoginRequest{UsernameOrEmail: "reader2", Password: "wrong-password"})
	if err == nil {
		t.Fatalf("expected error for invalid credentials")
	}
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUserNotFound(t *testing.T) {
	db := setupTestDB(t)

	_, err := Login(context.Background(), db, LoginRequest{UsernameOrEmail: "missing", Password: "password"})
	if err == nil {
		t.Fatalf("expected error for missing user")
	}
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
