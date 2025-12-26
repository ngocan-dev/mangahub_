package chat

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ngocan-dev/mangahub/backend/domain/friend"
)

var (
	ErrNotFriends        = errors.New("users are not friends")
	ErrEmptyMessage      = errors.New("message content is empty")
	ErrInvalidRoom       = errors.New("invalid chat room")
	ErrCannotMessageSelf = errors.New("cannot message yourself")
)

// Service orchestrates friend-only chat flows.
type Service struct {
	repo       *Repository
	friendRepo *friend.Repository
}

// NewService builds a chat service.
func NewService(repo *Repository, friendRepo *friend.Repository) *Service {
	return &Service{
		repo:       repo,
		friendRepo: friendRepo,
	}
}

// RoomCode builds the stable room code for two users.
func RoomCode(userA, userB int64) string {
	if userA > userB {
		userA, userB = userB, userA
	}
	return fmt.Sprintf("friend_%d_%d", userA, userB)
}

// SendMessage validates friendship, ensures a room exists, and saves the message.
func (s *Service) SendMessage(ctx context.Context, senderID, receiverID int64, content string) (*ChatRoom, *ChatMessage, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, nil, ErrEmptyMessage
	}
	if senderID == receiverID {
		return nil, nil, ErrCannotMessageSelf
	}

	areFriends, err := s.friendRepo.AreFriends(ctx, senderID, receiverID)
	if err != nil {
		return nil, nil, err
	}
	if !areFriends {
		return nil, nil, ErrNotFriends
	}

	room, err := s.repo.EnsurePrivateRoom(ctx, senderID, receiverID, senderID)
	if err != nil {
		return nil, nil, err
	}

	msg, err := s.repo.CreateMessage(ctx, room.ID, senderID, content)
	if err != nil {
		return nil, nil, err
	}

	return room, msg, nil
}

// ListMessages returns messages for a room if the requester is a participant.
func (s *Service) ListMessages(ctx context.Context, requesterID, roomID int64, limit, offset int) ([]ChatMessage, error) {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrInvalidRoom
	}

	u1, u2, parseErr := ParseRoomParticipants(room.Code)
	if parseErr != nil {
		return nil, parseErr
	}

	if requesterID != u1 && requesterID != u2 {
		return nil, ErrNotFriends
	}

	areFriends, err := s.friendRepo.AreFriends(ctx, u1, u2)
	if err != nil {
		return nil, err
	}
	if !areFriends {
		return nil, ErrNotFriends
	}

	return s.repo.ListMessages(ctx, room.ID, limit, offset)
}

// ListConversations returns accepted friends with their latest message/room.
func (s *Service) ListConversations(ctx context.Context, userID int64) ([]ConversationSummary, error) {
	return s.repo.ListConversations(ctx, userID)
}

// ParseRoomParticipants returns user ids encoded in the room code.
func ParseRoomParticipants(code string) (int64, int64, error) {
	trimmed := strings.TrimPrefix(code, "friend_")
	parts := strings.Split(trimmed, "_")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("%w: %s", ErrInvalidRoom, code)
	}

	u1, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %v", ErrInvalidRoom, err)
	}
	u2, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %v", ErrInvalidRoom, err)
	}
	return u1, u2, nil
}

// WithLastMessageAt is a helper to get the most recent timestamp for sorting on the frontend.
func WithLastMessageAt(conv ConversationSummary) time.Time {
	if conv.LastMessageAt != nil {
		return *conv.LastMessageAt
	}
	return time.Time{}
}
