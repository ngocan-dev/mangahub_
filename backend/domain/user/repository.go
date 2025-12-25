package user

import (
	"context"
	"database/sql"
	"errors"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	AvatarURL string `json:"avatar,omitempty"`
}

// Create inserts new user into SQLite DB
func (r *Repository) Create(username, email, passwordHash string) (int64, error) {
	res, err := r.db.Exec(`
        INSERT INTO Users (Username, Email, PasswordHash)
        VALUES (?, ?, ?)
    `, username, email, passwordHash)

	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetByID returns user by primary key
func (r *Repository) GetByID(id int64) (*User, error) {
	row := r.db.QueryRow(`
        SELECT UserId, Username, Email, PasswordHash
        FROM Users
        WHERE UserId = ?
    `, id)

	u := User{}
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password); err != nil {
		return nil, err
	}
	return &u, nil
}

// FindByUsernameOrEmail returns a user matching the provided username or email.
// It returns (nil, nil) when no user exists for the query.
func (r *Repository) FindByUsernameOrEmail(ctx context.Context, query string) (*User, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, username, email, password_hash, COALESCE(avatar_url, '')
        FROM users
        WHERE username = ? OR email = ?
        LIMIT 1
    `, query, query)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.AvatarURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
