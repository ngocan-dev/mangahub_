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
}

func (r *Repository) Create(username, email, passwordHash string) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO users (username, email, password_hash)
         VALUES (?, ?, ?)`,
		username, email, passwordHash,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetByID(id int64) (*User, error) {
	row := r.db.QueryRow(
		`SELECT id, username, email FROM users WHERE id = ?`,
		id,
	)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.Email); err != nil {
		return nil, err
	}
	return &u, nil
}
