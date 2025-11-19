package user

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type User struct {
	ID       int64
	Username string
	Email    string
	Password string
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
