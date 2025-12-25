package friend

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Repository handles friend-related persistence
type Repository struct {
	db *sql.DB
}

// NewRepository builds a friend repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindUsersByQuery searches users by username or email (case-insensitive) excluding self and existing friends.
func (r *Repository) FindUsersByQuery(ctx context.Context, userID int64, query string) ([]UserSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
WITH friend_ids AS (
    SELECT friend_id AS id FROM friends WHERE user_id = ?
    UNION
    SELECT user_id AS id FROM friends WHERE friend_id = ?
),
pending AS (
    SELECT CASE WHEN from_user_id = ? THEN to_user_id ELSE from_user_id END AS id
    FROM friend_requests
    WHERE status = 'pending' AND (from_user_id = ? OR to_user_id = ?)
)
SELECT id, username, email, COALESCE(avatar_url, '') AS avatar
FROM users
WHERE (LOWER(username) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?))
  AND id != ?
  AND id NOT IN (SELECT id FROM friend_ids)
  AND id NOT IN (SELECT id FROM pending)
LIMIT 20
    `, userID, userID, userID, userID, userID, "%"+query+"%", "%"+query+"%", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserSummary
	for rows.Next() {
		var user UserSummary
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.AvatarURL); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// FindUserByID returns a user summary by id
func (r *Repository) FindUserByID(ctx context.Context, id int64) (*UserSummary, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, username, email, COALESCE(avatar_url, '') as avatar
        FROM users
        WHERE id = ?
    `, id)

	var user UserSummary
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.AvatarURL); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// AreFriends returns true when two users have a bidirectional link
func (r *Repository) AreFriends(ctx context.Context, userID, friendID int64) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM friends WHERE user_id = ? AND friend_id = ?
    `, userID, friendID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasPendingRequest checks for an active pending request between users.
func (r *Repository) HasPendingRequest(ctx context.Context, fromUserID, toUserID int64) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM friend_requests
        WHERE status = 'pending' AND (
            (from_user_id = ? AND to_user_id = ?) OR
            (from_user_id = ? AND to_user_id = ?)
        )
    `, fromUserID, toUserID, toUserID, fromUserID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateFriendRequest inserts a pending friendship request
func (r *Repository) CreateFriendRequest(ctx context.Context, requesterID, targetID int64) (*FriendRequest, error) {
	res, err := r.db.ExecContext(ctx, `
        INSERT INTO friend_requests (from_user_id, to_user_id, status)
        VALUES (?, ?, 'pending')
    `, requesterID, targetID)
	if err != nil {
		return nil, err
	}
	requestID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetFriendRequestByID(ctx, requestID)
}

// GetFriendRequestByID fetches a friend request row.
func (r *Repository) GetFriendRequestByID(ctx context.Context, id int64) (*FriendRequest, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT fr.id, fr.from_user_id, fr.to_user_id, fr.status, fr.created_at, u.username
        FROM friend_requests fr
        JOIN users u ON u.id = fr.from_user_id
        WHERE fr.id = ?
    `, id)
	var fr FriendRequest
	if err := row.Scan(&fr.ID, &fr.FromUserID, &fr.ToUserID, &fr.Status, &fr.CreatedAt, &fr.FromUsername); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &fr, nil
}

// UpdateFriendRequestStatus updates a request status.
func (r *Repository) UpdateFriendRequestStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE friend_requests
        SET status = ?
        WHERE id = ?
    `, status, id)
	return err
}

// CreateFriendshipBidirectional stores two rows (user->friend and friend->user).
func (r *Repository) CreateFriendshipBidirectional(ctx context.Context, userID, friendID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	insertStmt := `INSERT OR IGNORE INTO friends (user_id, friend_id) VALUES (?, ?)`
	if _, err = tx.ExecContext(ctx, insertStmt, userID, friendID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, insertStmt, friendID, userID); err != nil {
		return err
	}
	return nil
}

// ListFriends returns accepted friends with basic profile data.
func (r *Repository) ListFriends(ctx context.Context, userID int64) ([]UserSummary, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT u.id, u.username, u.email, COALESCE(u.avatar_url, '') as avatar
        FROM friends f
        JOIN users u ON u.id = f.friend_id
        WHERE f.user_id = ?
        ORDER BY u.username
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []UserSummary
	for rows.Next() {
		var u UserSummary
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.AvatarURL); err != nil {
			return nil, err
		}
		friends = append(friends, u)
	}
	return friends, rows.Err()
}

// CountMutualFriendships is a helper used by tests/debugging.
func (r *Repository) CountMutualFriendships(ctx context.Context, userID, friendID int64) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM friends
        WHERE (user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)
    `, userID, friendID, friendID, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count mutual friendships: %w", err)
	}
	return total, nil
}

// ListIncomingRequests returns pending requests sent to the given user.
func (r *Repository) ListIncomingRequests(ctx context.Context, userID int64) ([]FriendRequest, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT fr.id, fr.from_user_id, fr.to_user_id, fr.status, fr.created_at, u.username
        FROM friend_requests fr
        JOIN users u ON u.id = fr.from_user_id
        WHERE fr.to_user_id = ? AND fr.status = 'pending'
        ORDER BY fr.created_at DESC
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []FriendRequest
	for rows.Next() {
		var fr FriendRequest
		if err := rows.Scan(&fr.ID, &fr.FromUserID, &fr.ToUserID, &fr.Status, &fr.CreatedAt, &fr.FromUsername); err != nil {
			return nil, err
		}
		requests = append(requests, fr)
	}
	return requests, rows.Err()
}
