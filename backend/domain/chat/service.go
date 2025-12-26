package chat

import (
	"context"
	"errors"
	"fmt"
	"sort"
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
	friends, err := s.friendRepo.ListFriends(ctx, userID)
	if err != nil {
		return nil, err
	}
	if friends == nil {
		return []ConversationSummary{}, nil
	}

	conversations := make([]ConversationSummary, 0, len(friends))

	for _, f := range friends {
		roomCode := RoomCode(userID, f.ID)
		room, err := s.repo.GetRoomByCode(ctx, roomCode)
		if err != nil {
			return nil, err
		}

		var roomIDPtr *int64
		var lastMsg *ChatMessage

		if room != nil {
			roomIDPtr = &room.ID
			if msg, msgErr := s.repo.GetLastMessage(ctx, room.ID); msgErr != nil {
				return nil, msgErr
			} else {
				lastMsg = msg
			}
		} else {
			createdRoom, ensureErr := s.repo.EnsurePrivateRoom(ctx, userID, f.ID, userID)
			if ensureErr != nil {
				return nil, ensureErr
			}
			if createdRoom != nil {
				roomIDPtr = &createdRoom.ID
			}
		}

		var avatarPtr *string
		if f.AvatarURL != "" {
			avatarPtr = &f.AvatarURL
		}

		conv := ConversationSummary{
			FriendID:       f.ID,
			FriendUsername: f.Username,
			FriendAvatar:   avatarPtr,
			RoomID:         roomIDPtr,
		}
		if lastMsg != nil {
			conv.LastMessage = lastMsg
			conv.LastMessageAt = &lastMsg.CreatedAt
		}

		conversations = append(conversations, conv)
	}

	// Most recent conversations first
	sort.SliceStable(conversations, func(i, j int) bool {
		li := WithLastMessageAt(conversations[i])
		lj := WithLastMessageAt(conversations[j])
		if li.Equal(lj) {
			return conversations[i].FriendUsername < conversations[j].FriendUsername
		}
		return li.After(lj)
	})

	return conversations, nil
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
