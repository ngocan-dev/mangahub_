package friend

import (
	"context"
	"database/sql"
)

// Repository handles friend-related persistence
type Repository struct {
	db *sql.DB
}

// NewRepository builds a friend repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindUserByUsername returns a user summary by username (case-insensitive)
func (r *Repository) FindUserByUsername(ctx context.Context, username string) (*UserSummary, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT UserId, Username
        FROM Users
        WHERE lower(Username) = lower(?)
    `, username)

	var user UserSummary
	if err := row.Scan(&user.ID, &user.Username); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetFriendship retrieves an existing friendship or request between two users
func (r *Repository) GetFriendship(ctx context.Context, userID, friendID int64) (*Friendship, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT Friendship_Id, User_Id, Friend_Id, Status, Created_At, Accepted_At
        FROM Friends
        WHERE (User_Id = ? AND Friend_Id = ?) OR (User_Id = ? AND Friend_Id = ?)
    `, userID, friendID, friendID, userID)

	var friendship Friendship
	var acceptedAt sql.NullTime
	if err := row.Scan(
		&friendship.ID,
		&friendship.UserID,
		&friendship.FriendID,
		&friendship.Status,
		&friendship.CreatedAt,
		&acceptedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if acceptedAt.Valid {
		friendship.AcceptedAt = &acceptedAt.Time
	}

	return &friendship, nil
}

// CreateFriendRequest inserts a pending friendship request
func (r *Repository) CreateFriendRequest(ctx context.Context, requesterID, targetID int64) (*Friendship, error) {
	res, err := r.db.ExecContext(ctx, `
        INSERT INTO Friends (User_Id, Friend_Id, Status)
        VALUES (?, ?, 'pending')
    `, requesterID, targetID)
	if err != nil {
		return nil, err
	}

	requestID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, `
        SELECT Friendship_Id, User_Id, Friend_Id, Status, Created_At, Accepted_At
        FROM Friends
        WHERE Friendship_Id = ?
    `, requestID)

	var friendship Friendship
	var acceptedAt sql.NullTime
	if err := row.Scan(
		&friendship.ID,
		&friendship.UserID,
		&friendship.FriendID,
		&friendship.Status,
		&friendship.CreatedAt,
		&acceptedAt,
	); err != nil {
		return nil, err
	}

	if acceptedAt.Valid {
		friendship.AcceptedAt = &acceptedAt.Time
	}

	return &friendship, nil
}

// AcceptFriendRequest updates a pending request to accepted status
func (r *Repository) AcceptFriendRequest(ctx context.Context, requesterID, targetID int64) (*Friendship, error) {
	res, err := r.db.ExecContext(ctx, `
        UPDATE Friends
        SET Status = 'accepted', Accepted_At = CURRENT_TIMESTAMP
        WHERE User_Id = ? AND Friend_Id = ? AND Status = 'pending'
    `, requesterID, targetID)
	if err != nil {
		return nil, err
	}

	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return nil, nil
	}

	return r.GetFriendship(ctx, requesterID, targetID)
}
