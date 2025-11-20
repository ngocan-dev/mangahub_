package friend

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ngocan-dev/mangahub/manga-backend/internal/security"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrCannotFriendSelf = errors.New("cannot add yourself as friend")
	ErrAlreadyFriends   = errors.New("already friends")
	ErrRequestPending   = errors.New("friend request already pending")
	ErrBlocked          = errors.New("friendship blocked")
	ErrNoPendingRequest = errors.New("no pending friend request")
	ErrInvalidUsername  = errors.New("invalid username")
)

// Notifier sends optional friend notifications
type Notifier interface {
	NotifyFriendRequest(ctx context.Context, targetUserID int64, requesterUsername string) error
	NotifyFriendAccepted(ctx context.Context, requesterID int64, targetUsername string) error
}

type noopNotifier struct{}

func (noopNotifier) NotifyFriendRequest(ctx context.Context, targetUserID int64, requesterUsername string) error {
	return nil
}

func (noopNotifier) NotifyFriendAccepted(ctx context.Context, requesterID int64, targetUsername string) error {
	return nil
}

// Service orchestrates friend workflows
type Service struct {
	repo     *Repository
	notifier Notifier
}

// NewService builds a friend service
func NewService(repo *Repository, notifier Notifier) *Service {
	if notifier == nil {
		notifier = noopNotifier{}
	}
	return &Service{repo: repo, notifier: notifier}
}

// SearchUser finds a user by username for friend lookup
func (s *Service) SearchUser(ctx context.Context, username string) (*UserSummary, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, ErrInvalidUsername
	}

	// Input length limits are enforced
	// SQL injection attempts are blocked
	// Invalid data formats are rejected
	if err := security.ValidateUsername(username); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidUsername, err)
	}

	user, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// SendFriendRequest sends a pending friend invitation to the target user
func (s *Service) SendFriendRequest(ctx context.Context, requesterID int64, requesterUsername, targetUsername string) (*Friendship, error) {
	targetUser, err := s.SearchUser(ctx, targetUsername)
	if err != nil {
		return nil, err
	}

	if targetUser.ID == requesterID {
		return nil, ErrCannotFriendSelf
	}

	friendship, err := s.repo.GetFriendship(ctx, requesterID, targetUser.ID)
	if err != nil {
		return nil, err
	}

	if friendship != nil {
		switch friendship.Status {
		case "accepted":
			return nil, ErrAlreadyFriends
		case "blocked":
			return nil, ErrBlocked
		case "pending":
			return nil, ErrRequestPending
		}
	}

	created, err := s.repo.CreateFriendRequest(ctx, requesterID, targetUser.ID)
	if err != nil {
		return nil, err
	}

	_ = s.notifier.NotifyFriendRequest(ctx, targetUser.ID, requesterUsername)

	return created, nil
}

// AcceptFriendRequest approves a pending request initiated by requesterUsername
func (s *Service) AcceptFriendRequest(ctx context.Context, userID int64, accepterUsername, requesterUsername string) (*Friendship, error) {
	requester, err := s.SearchUser(ctx, requesterUsername)
	if err != nil {
		return nil, err
	}

	friendship, err := s.repo.GetFriendship(ctx, requester.ID, userID)
	if err != nil {
		return nil, err
	}

	if friendship == nil || friendship.Status != "pending" || friendship.UserID != requester.ID {
		return nil, ErrNoPendingRequest
	}

	updated, err := s.repo.AcceptFriendRequest(ctx, requester.ID, userID)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrNoPendingRequest
	}

	_ = s.notifier.NotifyFriendAccepted(ctx, requester.ID, accepterUsername)

	return updated, nil
}
