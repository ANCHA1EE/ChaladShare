package models

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrInvalidSelfAction = errors.New("cannot act on yourself")
	ErrAlreadyFriends    = errors.New("already friends")
	ErrNotFriends        = errors.New("not friends")
	//ErrRequestNotFound     = errors.New("friend request not found or already decided") // ไม่พบคำขอ หรือคำขอถูกตัดสินใจไปแล้ว
	//ErrNotYourRequestToAct = errors.New("not your request to act on")                  // ไม่ใช่คำขอที่คุณต้องตัดสินใจ
)

type FriendRequestStatus string

const (
	FRPending  FriendRequestStatus = "pending"
	FRAccepted FriendRequestStatus = "accepted"
	FRDeclined FriendRequestStatus = "declined"
)

func (s FriendRequestStatus) Valid() bool {
	switch s {
	case FRPending, FRAccepted, FRDeclined:
		return true
	default:
		return false
	}
}

type Follow struct {
	FollowerID      int       `json:"follower_user_id" db:"follower_user_id"`
	FollowedID      int       `json:"followed_user_id" db:"followed_user_id"`
	FollowCreatedAt time.Time `json:"follow_created_at" db:"follow_created_at"`
}

type FriendRequest struct {
	RequestID        int                 `json:"request_id"`
	RequesterUserID  int                 `json:"requester_user_id"`
	AddresseeUserID  int                 `json:"addressee_user_id"`
	RequestStatus    FriendRequestStatus `json:"request_status"`
	RequestCreatedAt time.Time           `json:"request_created_at"`
	DecidedAt        sql.NullTime        `json:"decided_at"`
}

type Friendship struct {
	UserID    int       `json:"user_id"`
	FriendID  int       `json:"friend_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateFollowRequest struct {
	FollowedUserID int `json:"followed_user_id" binding:"required"`
}

type SendFriendRequest struct {
	ToUserID int `json:"to_user_id" binding:"required"`
}

type IncomingReqItem struct {
	RequestID       int       `json:"request_id"`
	RequesterUserID int       `json:"requester_user_id"`
	RequestedAt     time.Time `json:"requested_at"`
	Username        string    `json:"username"`
	Avatar          string    `json:"avatar"`
}

type OutgoingReqItem struct {
	RequestID    int       `json:"request_id"`
	TargetUserID int       `json:"target_user_id"`
	RequestedAt  time.Time `json:"requested_at"`
	Username     string    `json:"username"`
	Avatar       string    `json:"avatar"`
}

type FriendItem struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	Avatar      string `json:"avatar"`
	IsFriend    bool   `json:"is_friend"`
	IsFollowing bool   `json:"is_following"`
}

type FollowUser struct {
	UserID      int    `json:"user_id"`
	Username    string `json:"username"`
	Avatar      string `json:"avatar"`
	IsFriend    bool   `json:"is_friend"`
	IsFollowing bool   `json:"is_following"`
}

/* 20-02 by ploy */

type UserSearchItem struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`

	IsFriend         bool   `json:"is_friend"`
	RequestID        int    `json:"request_id"`
	RequestStatus    string `json:"request_status"`
	RequestDirection string `json:"request_direction"`
}

/* 20-02 by ploy */

func (r *CreateFollowRequest) Validate(actorID int) error {
	if r.FollowedUserID == actorID {
		return ErrInvalidSelfAction
	}
	return nil
}

func OrderedPair(a, b int) (low, high int) {
	if a < b {
		return a, b
	}
	return b, a
}

func (r *SendFriendRequest) Validate(requesterID int) error {
	if r.ToUserID == requesterID {
		return ErrInvalidSelfAction
	}
	return nil
}
