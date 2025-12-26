package friend

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Repository handles friend-related persistence
type Repository struct {
	db                *sql.DB
	friendIDColumn    string
	friendHasStatus   bool
	friendSchemaOnce  sync.Once
	friendSchemaError error
}

// NewRepository builds a friend repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db, friendIDColumn: "friend_user_id"}
}

// FindUsersByQuery searches users by username or email (case-insensitive) excluding self and existing friends.
func (r *Repository) FindUsersByQuery(
	ctx context.Context,
	userID int64,
	query string,
) ([]UserSummary, error) {

	rows, err := r.db.QueryContext(ctx, `
SELECT
	id,
	username,
	email,
	COALESCE(avatar_url, '') AS avatar_url
FROM users
WHERE
	id != ?
	AND (
		LOWER(username) LIKE LOWER(?)
		OR LOWER(email) LIKE LOWER(?)
	)
LIMIT 20
	`, userID, "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserSummary
	for rows.Next() {
		var u UserSummary
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.AvatarURL); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

// FindUserByID returns a user summary by id
func (r *Repository) FindUserByID(ctx context.Context, id int64) (*UserSummary, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, username, email, COALESCE(avatar_url, '') as avatar_url
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
	if err := r.ensureFriendSchema(ctx); err != nil {
		return false, err
	}

	query := fmt.Sprintf(`
		SELECT 1
		FROM friends
		WHERE status = 'accepted'
		  AND (
			  (user_id = ? AND %s = ?)
		   OR (user_id = ? AND %s = ?)
		  )
		LIMIT 1
	`, r.friendIDColumn, r.friendIDColumn)

	var flag int
	if err := r.db.QueryRowContext(
		ctx,
		query,
		userID,
		friendID,
		friendID,
		userID,
	).Scan(&flag); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// HasPendingRequest checks for an active pending request between users.
func (r *Repository) HasPendingRequest(ctx context.Context, fromUserID, toUserID int64) (bool, error) {
	if err := r.ensureFriendSchema(ctx); err != nil {
		return false, err
	}

	var count int
	query := fmt.Sprintf(`
        SELECT COUNT(*) FROM friends
        WHERE status = 'pending' AND (
            (user_id = ? AND %s = ?) OR
            (user_id = ? AND %s = ?)
        )
    `, r.friendIDColumn, r.friendIDColumn)
	if err := r.db.QueryRowContext(ctx, query, fromUserID, toUserID, toUserID, fromUserID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateFriendRequest inserts a pending friendship request
func (r *Repository) CreateFriendRequest(ctx context.Context, requesterID, targetID int64) (*FriendRequest, error) {
	if err := r.ensureFriendSchema(ctx); err != nil {
		return nil, err
	}

	insertStmt := fmt.Sprintf(`
        INSERT INTO friends (user_id, %s, status)
        VALUES (?, ?, 'pending')
        ON DUPLICATE KEY UPDATE status = VALUES(status)
    `, r.friendIDColumn)
	res, err := r.db.ExecContext(ctx, insertStmt, requesterID, targetID)
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
	if err := r.ensureFriendSchema(ctx); err != nil {
		return nil, err
	}

	row := r.db.QueryRowContext(ctx, `
        SELECT fr.id, fr.user_id, fr.`+r.friendIDColumn+`, fr.status, fr.created_at, u.username
        FROM friends fr
        JOIN users u ON u.id = fr.user_id
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
	if err := r.ensureFriendSchema(ctx); err != nil {
		return err
	}

	res, err := r.db.ExecContext(ctx, `
        UPDATE friends
        SET status = ?
        WHERE id = ?
    `, status, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// CreateFriendshipBidirectional stores two rows (user->friend and friend->user).
func (r *Repository) CreateFriendshipBidirectional(ctx context.Context, userID, friendID int64) error {
	if err := r.ensureFriendSchema(ctx); err != nil {
		return err
	}

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

	insertStmt := fmt.Sprintf(`
        INSERT INTO friends (user_id, %s, status)
        VALUES (?, ?, 'accepted')
        ON DUPLICATE KEY UPDATE status = VALUES(status)
    `, r.friendIDColumn)

	if _, err = tx.ExecContext(ctx, insertStmt, userID, friendID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, insertStmt, friendID, userID); err != nil {
		return err
	}
	return nil
}

// AcceptFriendRequestTx atomically accepts the pending request and ensures bidirectional accepted rows.
func (r *Repository) AcceptFriendRequestTx(ctx context.Context, requestID, fromUserID, toUserID int64) error {
	if err := r.ensureFriendSchema(ctx); err != nil {
		return err
	}

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

	updateRes, err := tx.ExecContext(ctx, `
        UPDATE friends
        SET status = 'accepted'
        WHERE id = ? AND user_id = ? AND `+r.friendIDColumn+` = ? AND status = 'pending'
    `, requestID, fromUserID, toUserID)
	if err != nil {
		return err
	}
	affected, err := updateRes.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	insertStmt := fmt.Sprintf(`
        INSERT INTO friends (user_id, %s, status)
        VALUES (?, ?, 'accepted')
        ON DUPLICATE KEY UPDATE status = VALUES(status)
    `, r.friendIDColumn)

	if _, err = tx.ExecContext(ctx, insertStmt, fromUserID, toUserID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, insertStmt, toUserID, fromUserID); err != nil {
		return err
	}

	return nil
}

// ListFriends returns accepted friends with basic profile data.
func (r *Repository) ListFriends(ctx context.Context, userID int64) ([]UserSummary, error) {
	if err := r.ensureFriendSchema(ctx); err != nil {
		return nil, err
	}

	statusFilter := ""
	if r.friendHasStatus {
		statusFilter = " AND status = 'accepted'"
	}

	query := fmt.Sprintf(`
        SELECT u.id, u.username, u.email, COALESCE(u.avatar_url, '') as avatar
        FROM (
            SELECT %[1]s AS friend_id FROM friends WHERE user_id = ?%[2]s
            UNION
            SELECT user_id AS friend_id FROM friends WHERE %[1]s = ?%[2]s
        ) rel
        JOIN users u ON u.id = rel.friend_id
        ORDER BY u.username
    `, r.friendIDColumn, statusFilter)

	rows, err := r.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		if r.isUnknownColumnError(err, r.friendIDColumn) {
			return nil, fmt.Errorf("friends table is missing expected column %q", r.friendIDColumn)
		}
		return nil, err
	}
	defer rows.Close()

	return scanUserSummaries(rows)
}

// CountMutualFriendships is a helper used by tests/debugging.
func (r *Repository) CountMutualFriendships(ctx context.Context, userID, friendID int64) (int, error) {
	if err := r.ensureFriendSchema(ctx); err != nil {
		return 0, err
	}

	var total int
	query := fmt.Sprintf(`
        SELECT COUNT(*) FROM friends
        WHERE ((user_id = ? AND %s = ?) OR (user_id = ? AND %s = ?))`, r.friendIDColumn, r.friendIDColumn)
	if r.friendHasStatus {
		query += " AND status = 'accepted'"
	}
	err := r.db.QueryRowContext(ctx, query, userID, friendID, friendID, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count mutual friendships: %w", err)
	}
	return total, nil
}

// ListIncomingRequests returns pending requests sent to the given user.
func (r *Repository) ListIncomingRequests(ctx context.Context, userID int64) ([]FriendRequest, error) {
	if err := r.ensureFriendSchema(ctx); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
        SELECT fr.id, fr.user_id, fr.`+r.friendIDColumn+`, fr.status, fr.created_at, u.username
        FROM friends fr
        JOIN users u ON u.id = fr.user_id
        WHERE fr.`+r.friendIDColumn+` = ? AND fr.status = 'pending'
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

// FindFriendshipBetween returns a friendship row (any status) between two users if it exists.
func (r *Repository) FindFriendshipBetween(ctx context.Context, userID, friendID int64) (*Friendship, error) {
	if err := r.ensureFriendSchema(ctx); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
        SELECT id, user_id, %s, status, created_at
        FROM friends
        WHERE (user_id = ? AND %s = ?) OR (user_id = ? AND %s = ?)
        LIMIT 1
    `, r.friendIDColumn, r.friendIDColumn, r.friendIDColumn)

	row := r.db.QueryRowContext(ctx, query, userID, friendID, friendID, userID)
	var fr Friendship
	if err := row.Scan(&fr.ID, &fr.UserID, &fr.FriendID, &fr.Status, &fr.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &fr, nil
}

func (r *Repository) ensureFriendSchema(ctx context.Context) error {
	r.friendSchemaOnce.Do(func() {
		if err := r.detectFriendIDColumn(ctx); err != nil {
			r.friendSchemaError = err
			return
		}

		// Detect optional status column to filter accepted friends
		if _, err := r.db.ExecContext(ctx, `SELECT status FROM friends LIMIT 0`); err == nil {
			r.friendHasStatus = true
		}
	})

	return r.friendSchemaError
}

func (r *Repository) detectFriendIDColumn(ctx context.Context) error {
	query := `SELECT friend_user_id FROM friends LIMIT 0`
	if _, err := r.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("friends table must have friend_user_id column")
	}
	r.friendIDColumn = "friend_user_id"
	return nil
}

func (r *Repository) isUnknownColumnError(err error, column string) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	col := strings.ToLower(column)
	return (strings.Contains(msg, "no such column") ||
		strings.Contains(msg, "unknown column") ||
		strings.Contains(msg, "does not exist")) && strings.Contains(msg, col)
}

func scanUserSummaries(rows *sql.Rows) ([]UserSummary, error) {
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
