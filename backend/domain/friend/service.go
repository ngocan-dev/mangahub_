package friend

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ngocan-dev/mangahub/backend/domain/user"
	"github.com/ngocan-dev/mangahub/backend/internal/security"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrCannotFriendSelf = errors.New("cannot add yourself as friend")
	ErrAlreadyFriends   = errors.New("already friends")
	ErrRequestPending   = errors.New("friend request already pending")
	ErrRequestRejected  = errors.New("friend request rejected")
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
	userRepo *user.Repository
	notifier Notifier
}

// NewService builds a friend service
func NewService(repo *Repository, userRepo *user.Repository, notifier Notifier) *Service {
	if notifier == nil {
		notifier = noopNotifier{}
	}
	return &Service{repo: repo, userRepo: userRepo, notifier: notifier}
}

// SearchUsers looks up users by username or email, tolerating empty datasets.
func (s *Service) SearchUsers(ctx context.Context, userID int64, query string) ([]UserSummary, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, ErrInvalidUsername
	}

	users, err := s.repo.FindUsersByQuery(ctx, userID, query)
	if err != nil {
		return nil, err
	}
	if users == nil {
		return []UserSummary{}, nil
	}
	return users, nil
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

	user, err := s.userRepo.FindByUsernameOrEmail(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return &UserSummary{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		AvatarURL: user.AvatarURL,
	}, nil
}

// SendFriendRequest sends a pending friend invitation to the target user id.
func (s *Service) SendFriendRequest(ctx context.Context, requesterID int64, requesterUsername string, targetUserID int64) (*FriendRequest, error) {
	targetUser, err := s.repo.FindUserByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if targetUser == nil {
		return nil, ErrUserNotFound
	}
	if targetUser.ID == requesterID {
		return nil, ErrCannotFriendSelf
	}

	alreadyFriends, err := s.repo.AreFriends(ctx, requesterID, targetUser.ID)
	if err != nil {
		return nil, err
	}
	if alreadyFriends {
		return nil, ErrAlreadyFriends
	}

	existing, err := s.repo.FindFriendshipBetween(ctx, requesterID, targetUser.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		switch existing.Status {
		case "pending":
			return nil, ErrRequestPending
		case "accepted":
			return nil, ErrAlreadyFriends
		case "blocked":
			return nil, ErrBlocked
		}
	}

	created, err := s.repo.CreateFriendRequest(ctx, requesterID, targetUser.ID)
	if err != nil {
		return nil, err
	}

	_ = s.notifier.NotifyFriendRequest(ctx, targetUser.ID, requesterUsername)

	return created, nil
}

// AcceptFriendRequest approves a pending request and creates friendship rows.
func (s *Service) AcceptFriendRequest(ctx context.Context, userID int64, accepterUsername string, requestID int64) (*Friendship, error) {
	req, err := s.repo.GetFriendRequestByID(ctx, requestID)
	if err != nil {
		return nil, err
	}
	if req == nil || req.Status != "pending" || req.ToUserID != userID {
		return nil, ErrNoPendingRequest
	}

	if err := s.repo.AcceptFriendRequestTx(ctx, req.ID, req.FromUserID, req.ToUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoPendingRequest
		}
		return nil, err
	}

	now := time.Now()
	friendship := &Friendship{
		UserID:     req.FromUserID,
		FriendID:   req.ToUserID,
		Status:     "accepted",
		CreatedAt:  req.CreatedAt,
		AcceptedAt: &now,
	}

	_ = s.notifier.NotifyFriendAccepted(ctx, req.FromUserID, accepterUsername)

	return friendship, nil
}

// RejectFriendRequest marks a pending request as rejected.
func (s *Service) RejectFriendRequest(ctx context.Context, userID, requestID int64) error {
	req, err := s.repo.GetFriendRequestByID(ctx, requestID)
	if err != nil {
		return err
	}
	if req == nil || req.ToUserID != userID || req.Status != "pending" {
		return ErrNoPendingRequest
	}
	if err := s.repo.UpdateFriendRequestStatus(ctx, requestID, "blocked"); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoPendingRequest
		}
		return err
	}
	return nil
}

// ListFriends returns accepted friends for a user.
func (s *Service) ListFriends(ctx context.Context, userID int64) ([]UserSummary, error) {
	friends, err := s.repo.ListFriends(ctx, userID)
	if err != nil {
		return nil, err
	}
	if friends == nil {
		return []UserSummary{}, nil
	}
	return friends, nil
}

// ListPendingRequests returns pending requests directed to the user.
func (s *Service) ListPendingRequests(ctx context.Context, userID int64) ([]FriendRequest, error) {
	requests, err := s.repo.ListIncomingRequests(ctx, userID)
	if err != nil {
		return nil, err
	}
	if requests == nil {
		return []FriendRequest{}, nil
	}
	return requests, nil
}
