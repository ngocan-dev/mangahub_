package chat

import (
	"context"
	"database/sql"

	"github.com/go-sql-driver/mysql"
)

// Repository handles chat persistence.
type Repository struct {
	db *sql.DB
}

// NewRepository constructs a chat repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

//
// -------------------------
// ROOM
// -------------------------
//

// EnsurePrivateRoom returns the existing private room for a friend pair
// or creates it if missing.
// IMPORTANT: This must ONLY be called when user clicks a friend.
func (r *Repository) EnsurePrivateRoom(
	ctx context.Context,
	userA, userB, createdBy int64,
) (*ChatRoom, error) {

	code := RoomCode(userA, userB)

	// 1. Try existing room
	room, err := r.GetRoomByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if room != nil {
		return room, nil
	}

	// 2. Create new room
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO chat_rooms (code, name, is_private, created_by)
		VALUES (?, ?, TRUE, ?)
	`, code, "Private Chat", createdBy)
	if err != nil {
		// Handle duplicate key (room created concurrently)
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return r.GetRoomByCode(ctx, code)
		}
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
	if err := row.Scan(
		&room.ID,
		&room.Code,
		&room.Name,
		&room.IsPrivate,
		&room.CreatedBy,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &room, nil
}

// GetRoomByCode fetches a chat room by its stable code.
func (r *Repository) GetRoomByCode(ctx context.Context, code string) (*ChatRoom, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, code, COALESCE(name, ''), is_private, created_by
		FROM chat_rooms
		WHERE code = ?
		LIMIT 1
	`, code)

	var room ChatRoom
	if err := row.Scan(
		&room.ID,
		&room.Code,
		&room.Name,
		&room.IsPrivate,
		&room.CreatedBy,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &room, nil
}

//
// -------------------------
// MESSAGES
// -------------------------
//

// CreateMessage persists a chat message.
func (r *Repository) CreateMessage(
	ctx context.Context,
	roomID, userID int64,
	content string,
) (*ChatMessage, error) {

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
	if err := row.Scan(
		&msg.ID,
		&msg.RoomID,
		&msg.UserID,
		&msg.Content,
		&msg.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &msg, nil
}

// ListMessages returns messages ordered ascending by creation time.
func (r *Repository) ListMessages(
	ctx context.Context,
	roomID int64,
	limit, offset int,
) ([]ChatMessage, error) {

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
		if err := rows.Scan(
			&msg.ID,
			&msg.RoomID,
			&msg.UserID,
			&msg.Content,
			&msg.CreatedAt,
		); err != nil {
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
	if err := row.Scan(
		&msg.ID,
		&msg.RoomID,
		&msg.UserID,
		&msg.Content,
		&msg.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &msg, nil
}

//
// -------------------------
// CONVERSATIONS (READ ONLY)
// -------------------------
//

// ListConversations returns accepted friends with optional room + last message.
// IMPORTANT: This function MUST NOT create rooms.
func (r *Repository) ListConversations(
	ctx context.Context,
	userID int64,
) ([]ConversationSummary, error) {

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			u.id AS friend_id,
			u.username AS friend_username,
			COALESCE(u.avatar_url, '') AS friend_avatar,
			cr.id AS room_id,
			cm.id AS last_message_id,
			cm.room_id AS last_message_room_id,
			cm.user_id AS last_message_user_id,
			cm.content AS last_message_content,
			cm.created_at AS last_message_created_at
		FROM (
			SELECT friend_user_id AS friend_id
			FROM friends
			WHERE user_id = ? AND status = 'accepted'
			UNION
			SELECT user_id AS friend_id
			FROM friends
			WHERE friend_user_id = ? AND status = 'accepted'
		) rel
		JOIN users u ON u.id = rel.friend_id
		LEFT JOIN chat_rooms cr
		  ON cr.code = CONCAT('friend_', LEAST(?, u.id), '_', GREATEST(?, u.id))
		 AND cr.is_private = TRUE
		LEFT JOIN chat_messages cm
		  ON cm.id = (
			SELECT id
			FROM chat_messages
			WHERE room_id = cr.id
			ORDER BY created_at DESC, id DESC
			LIMIT 1
		  )
		ORDER BY
			COALESCE(cm.created_at, '1970-01-01 00:00:00') DESC,
			u.username ASC
	`, userID, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []ConversationSummary

	for rows.Next() {
		var (
			friendID       int64
			friendUsername string
			friendAvatar   string
			roomID         sql.NullInt64
			lastMsgID      sql.NullInt64
			lastMsgRoomID  sql.NullInt64
			lastMsgUserID  sql.NullInt64
			lastMsgContent sql.NullString
			lastMsgCreated sql.NullTime
		)

		if err := rows.Scan(
			&friendID,
			&friendUsername,
			&friendAvatar,
			&roomID,
			&lastMsgID,
			&lastMsgRoomID,
			&lastMsgUserID,
			&lastMsgContent,
			&lastMsgCreated,
		); err != nil {
			return nil, err
		}

		var avatarPtr *string
		if friendAvatar != "" {
			avatarPtr = &friendAvatar
		}

		var roomIDPtr *int64
		if roomID.Valid {
			roomIDPtr = &roomID.Int64
		}

		conv := ConversationSummary{
			FriendID:       friendID,
			FriendUsername: friendUsername,
			FriendAvatar:   avatarPtr,
			RoomID:         roomIDPtr,
		}

		if lastMsgID.Valid && lastMsgRoomID.Valid &&
			lastMsgUserID.Valid && lastMsgCreated.Valid {

			lastMsg := ChatMessage{
				ID:        lastMsgID.Int64,
				RoomID:    lastMsgRoomID.Int64,
				UserID:    lastMsgUserID.Int64,
				Content:   lastMsgContent.String,
				CreatedAt: lastMsgCreated.Time,
			}
			conv.LastMessage = &lastMsg
			conv.LastMessageAt = &lastMsg.CreatedAt
		}

		conversations = append(conversations, conv)
	}

	return conversations, rows.Err()
}
