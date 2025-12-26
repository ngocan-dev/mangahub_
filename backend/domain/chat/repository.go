package chat

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository handles chat persistence.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs a chat repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// EnsurePrivateRoom returns the existing room for a friend pair or creates it if missing.
func (r *Repository) EnsurePrivateRoom(ctx context.Context, userA, userB, createdBy int64) (*ChatRoom, error) {
	code := RoomCode(userA, userB)
	name := fmt.Sprintf("Direct chat %d-%d", userA, userB)

	res, err := r.db.ExecContext(ctx, `
        INSERT INTO chat_rooms (code, name, is_private, created_by)
        VALUES (?, ?, TRUE, ?)
        ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)
    `, code, name, createdBy)
	if err != nil {
		return nil, err
	}

	roomID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetRoomByID(ctx, roomID)
}

// GetRoomByID fetches a chat room by id.
func (r *Repository) GetRoomByID(ctx context.Context, roomID int64) (*ChatRoom, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, code, COALESCE(name, ''), is_private, created_by
        FROM chat_rooms
        WHERE id = ?
    `, roomID)

	var room ChatRoom
	if err := row.Scan(&room.ID, &room.Code, &room.Name, &room.IsPrivate, &room.CreatedBy); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

// CreateMessage persists a chat message.
func (r *Repository) CreateMessage(ctx context.Context, roomID, userID int64, content string) (*ChatMessage, error) {
	res, err := r.db.ExecContext(ctx, `
        INSERT INTO chat_messages (room_id, user_id, content, created_at)
        VALUES (?, ?, ?, NOW())
    `, roomID, userID, content)
	if err != nil {
		return nil, err
	}
	msgID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetMessageByID(ctx, msgID)
}

// GetMessageByID fetches a single message.
func (r *Repository) GetMessageByID(ctx context.Context, id int64) (*ChatMessage, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, room_id, user_id, content, created_at
        FROM chat_messages
        WHERE id = ?
    `, id)
	var msg ChatMessage
	if err := row.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.Content, &msg.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

// ListMessages returns messages ordered ascending by creation time.
func (r *Repository) ListMessages(ctx context.Context, roomID int64, limit, offset int) ([]ChatMessage, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
        SELECT id, room_id, user_id, content, created_at
        FROM chat_messages
        WHERE room_id = ?
        ORDER BY created_at ASC, id ASC
        LIMIT ? OFFSET ?
    `, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []ChatMessage
	for rows.Next() {
		var msg ChatMessage
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

// GetLastMessage returns the most recent message for a room, if any.
func (r *Repository) GetLastMessage(ctx context.Context, roomID int64) (*ChatMessage, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT id, room_id, user_id, content, created_at
        FROM chat_messages
        WHERE room_id = ?
        ORDER BY created_at DESC, id DESC
        LIMIT 1
    `, roomID)

	var msg ChatMessage
	if err := row.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.Content, &msg.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

// ListConversations returns accepted friends with optional room + last message.
func (r *Repository) ListConversations(ctx context.Context, userID int64) ([]ConversationSummary, error) {
	query := `
        WITH accepted_friends AS (
            SELECT DISTINCT
                CASE
                    WHEN f.user_id = ? THEN f.friend_user_id
                    ELSE f.user_id
                END AS friend_id
            FROM friends f
            WHERE f.status = 'accepted'
              AND (f.user_id = ? OR f.friend_user_id = ?)
        )
        SELECT
            u.id AS friend_id,
            u.username AS friend_username,
            u.avatar_url AS friend_avatar,
            cr.id AS room_id,
            lm.content AS last_message,
            lm.created_at AS last_message_at
        FROM accepted_friends af
        JOIN users u ON u.id = af.friend_id
        LEFT JOIN chat_rooms cr
            ON cr.is_private = TRUE
           AND cr.code = CONCAT('friend_', LEAST(?, u.id), '_', GREATEST(?, u.id))
           AND cr.created_by IN (?, u.id)
        LEFT JOIN chat_messages lm ON lm.id = (
            SELECT cm.id
            FROM chat_messages cm
            WHERE cm.room_id = cr.id
            ORDER BY cm.created_at DESC, cm.id DESC
            LIMIT 1
        )
        ORDER BY lm.created_at DESC IS NULL, lm.created_at DESC, u.username
    `

	rows, err := r.db.QueryContext(ctx, query, userID, userID, userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []ConversationSummary
	for rows.Next() {
		var conv ConversationSummary
		var roomID sql.NullInt64
		var friendAvatar sql.NullString
		var lastMessage sql.NullString
		var lastMessageAt sql.NullTime

		if err := rows.Scan(
			&conv.FriendID,
			&conv.FriendUsername,
			&friendAvatar,
			&roomID,
			&lastMessage,
			&lastMessageAt,
		); err != nil {
			return nil, err
		}

		if roomID.Valid {
			conv.RoomID = &roomID.Int64
		}

		if friendAvatar.Valid {
			conv.FriendAvatar = &friendAvatar.String
		}

		if lastMessage.Valid {
			conv.LastMessage = &lastMessage.String
		}
		if lastMessageAt.Valid {
			conv.LastMessageAt = &lastMessageAt.Time
		}

		conversations = append(conversations, conv)
	}

	return conversations, rows.Err()
}
