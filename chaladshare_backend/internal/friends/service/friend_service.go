package service

import (
	"context"
	"errors"
	"strings"

	"chaladshare_backend/internal/friends/models"
	"chaladshare_backend/internal/friends/repository"
)

var (
	ErrBadRequest = errors.New("bad request")
	ErrForbidden  = errors.New("forbidden")
)

type FriendService interface {
	//Follow
	FollowUser(ctx context.Context, actorID, targetID int) error
	UnfollowUser(ctx context.Context, actorID, targetID int) error
	IsFollowing(ctx context.Context, actorID, targetID int) (bool, error)

	AreFriends(ctx context.Context, aID, bID int) (bool, error)

	// Lists (มี guard และ pagination ใน service) =====
	ListFriends(ctx context.Context, viewerID, userID int, search string, page, size int) ([]models.FriendItem, int, error)
	ListFollowers(ctx context.Context, viewerID, userID int, search string, page, size int) ([]models.FollowUser, int, error) // owner-only
	ListFollowing(ctx context.Context, viewerID, userID int, search string, page, size int) ([]models.FollowUser, int, error) // owner-only

	// Counts
	GetFollowStats(ctx context.Context, userID int) (followers int, following int, friends int, err error)

	// Friend Requests
	SendFriendRequest(ctx context.Context, actorID, toUserID int) (requestID int, err error)
	ListIncomingRequests(ctx context.Context, actorID int, page, size int) ([]models.IncomingReqItem, int, error)
	ListOutgoingRequests(ctx context.Context, actorID int, page, size int) ([]models.OutgoingReqItem, int, error)
	AcceptFriendRequest(ctx context.Context, actorID, requestID int) error
	DeclineFriendRequest(ctx context.Context, actorID, requestID int) error
	CancelFriendRequest(ctx context.Context, actorID, requestID int) error
	Unfriend(ctx context.Context, actorID, otherID int) error
	/* 20-02 by ploy */
	SearchAddFriend(ctx context.Context, actorID int, search string, page, size int) ([]models.UserSearchItem, int, error)
	/* 20-02 by ploy */

}

type friendsService struct {
	friendsrepo repository.FriendRepository
}

func NewFriendService(friendsrepo repository.FriendRepository) FriendService {
	return &friendsService{friendsrepo: friendsrepo}
}

func normalizeSearch(s string) string {
	return strings.TrimSpace(s)
}

func clampPageSize(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return page, size
}

func toLimitOffset(page, size int) (int, int) {
	return size, (page - 1) * size
}

func (s *friendsService) FollowUser(ctx context.Context, actorID, targetID int) error {
	if actorID == 0 || targetID == 0 {
		return ErrBadRequest
	}
	if actorID == targetID {
		return models.ErrInvalidSelfAction
	}
	return s.friendsrepo.InsertFollow(ctx, actorID, targetID)
}

func (s *friendsService) UnfollowUser(ctx context.Context, actorID, targetID int) error {
	if actorID == 0 || targetID == 0 {
		return ErrBadRequest
	}
	return s.friendsrepo.DeleteFollow(ctx, actorID, targetID)
}

func (s *friendsService) IsFollowing(ctx context.Context, actorID, targetID int) (bool, error) {
	if actorID == 0 || targetID == 0 {
		return false, ErrBadRequest
	}
	return s.friendsrepo.Following(ctx, actorID, targetID)
}

func (s *friendsService) AreFriends(ctx context.Context, aID, bID int) (bool, error) {
	if aID == 0 || bID == 0 {
		return false, ErrBadRequest
	}
	return s.friendsrepo.AreFriends(ctx, aID, bID)
}

func (s *friendsService) ListFriends(ctx context.Context, viewerID, userID int, search string, page, size int) ([]models.FriendItem, int, error) {
	if userID == 0 {
		return nil, 0, ErrBadRequest
	}
	search = normalizeSearch(search)
	page, size = clampPageSize(page, size)
	limit, offset := toLimitOffset(page, size)

	items, err := s.friendsrepo.ListFriends(ctx, viewerID, userID, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.friendsrepo.CountFriends(ctx, userID)
	return items, total, err
}

func (s *friendsService) ListFollowers(ctx context.Context, viewerID, userID int, search string, page, size int) ([]models.FollowUser, int, error) {
	// owner-only: คนอื่นเห็นเฉพาะตัวเลข
	if viewerID == 0 || userID == 0 {
		return nil, 0, ErrBadRequest
	}
	if viewerID != userID {
		return nil, 0, ErrForbidden
	}

	search = normalizeSearch(search)
	page, size = clampPageSize(page, size)
	limit, offset := toLimitOffset(page, size)

	items, err := s.friendsrepo.ListFollowers(ctx, viewerID, userID, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.friendsrepo.CountFollowers(ctx, userID)
	return items, total, err
}

func (s *friendsService) ListFollowing(ctx context.Context, viewerID, userID int, search string, page, size int) ([]models.FollowUser, int, error) {
	// owner-only: คนอื่นเห็นเฉพาะตัวเลข
	if viewerID == 0 || userID == 0 {
		return nil, 0, ErrBadRequest
	}
	if viewerID != userID {
		return nil, 0, ErrForbidden
	}

	search = normalizeSearch(search)
	page, size = clampPageSize(page, size)
	limit, offset := toLimitOffset(page, size)

	items, err := s.friendsrepo.ListFollowing(ctx, viewerID, userID, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.friendsrepo.CountFollowing(ctx, userID)
	return items, total, err
}

func (s *friendsService) GetFollowStats(ctx context.Context, userID int) (followers int, following int, friends int, err error) {
	if userID == 0 {
		return 0, 0, 0, ErrBadRequest
	}
	followers, err = s.friendsrepo.CountFollowers(ctx, userID)
	if err != nil {
		return
	}
	following, err = s.friendsrepo.CountFollowing(ctx, userID)
	if err != nil {
		return
	}
	friends, err = s.friendsrepo.CountFriends(ctx, userID)
	return
}

func (s *friendsService) SendFriendRequest(ctx context.Context, actorID, toUserID int) (int, error) {
	if actorID == 0 || toUserID == 0 {
		return 0, ErrBadRequest
	}
	// กันส่งหาตัวเอง
	if actorID == toUserID {
		return 0, models.ErrInvalidSelfAction
	}
	// กันกรณีเป็นเพื่อนกันอยู่แล้ว
	isFriend, err := s.friendsrepo.AreFriends(ctx, actorID, toUserID)
	if err != nil {
		return 0, err
	}
	if isFriend {
		return 0, models.ErrAlreadyFriends
	}
	// กัน pending ซ้ำ
	hasPending, err := s.friendsrepo.GetPendingBetween(ctx, actorID, toUserID)
	if err != nil {
		return 0, err
	}
	if hasPending {
		return 0, ErrBadRequest
	}

	return s.friendsrepo.CreateFriendRequest(ctx, actorID, toUserID)
}

func (s *friendsService) ListIncomingRequests(ctx context.Context, actorID int, page, size int) ([]models.IncomingReqItem, int, error) {
	if actorID == 0 {
		return nil, 0, ErrBadRequest
	}
	page, size = clampPageSize(page, size)
	limit, offset := toLimitOffset(page, size)

	items, err := s.friendsrepo.ListIncomingRequests(ctx, actorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total := len(items)
	return items, total, nil
}

func (s *friendsService) ListOutgoingRequests(ctx context.Context, actorID int, page, size int) ([]models.OutgoingReqItem, int, error) {
	if actorID == 0 {
		return nil, 0, ErrBadRequest
	}
	page, size = clampPageSize(page, size)
	limit, offset := toLimitOffset(page, size)

	items, err := s.friendsrepo.ListOutgoingRequests(ctx, actorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	// เช่นเดียวกับ incoming — ถ้าอยากได้ total จริง ให้เพิ่มเมธอด CountOutgoingPending
	total := len(items)
	return items, total, nil
}

func (s *friendsService) AcceptFriendRequest(ctx context.Context, actorID, requestID int) error {
	if actorID == 0 || requestID == 0 {
		return ErrBadRequest
	}
	// ตรวจสิทธิ์ว่า actor ต้องเป็น "ผู้รับ"
	fr, err := s.friendsrepo.GetFriendRequest(ctx, requestID)
	if err != nil {
		return err
	}
	if fr.AddresseeUserID != actorID || fr.RequestStatus != models.FRPending {
		return ErrForbidden
	}
	return s.friendsrepo.AcceptFriendRequest(ctx, requestID, actorID)
}

func (s *friendsService) DeclineFriendRequest(ctx context.Context, actorID, requestID int) error {
	if actorID == 0 || requestID == 0 {
		return ErrBadRequest
	}
	fr, err := s.friendsrepo.GetFriendRequest(ctx, requestID)
	if err != nil {
		return err
	}
	if fr.AddresseeUserID != actorID || fr.RequestStatus != models.FRPending {
		return ErrForbidden
	}
	return s.friendsrepo.DeclineFriendRequest(ctx, requestID, actorID)
}

func (s *friendsService) CancelFriendRequest(ctx context.Context, actorID, requestID int) error {
	if actorID == 0 || requestID == 0 {
		return ErrBadRequest
	}
	fr, err := s.friendsrepo.GetFriendRequest(ctx, requestID)
	if err != nil {
		return err
	}
	if fr.RequesterUserID != actorID || fr.RequestStatus != models.FRPending {
		return ErrForbidden
	}
	return s.friendsrepo.CancelFriendRequest(ctx, requestID, actorID)
}

func (s *friendsService) Unfriend(ctx context.Context, actorID, otherID int) error {
	if actorID == 0 || otherID == 0 {
		return ErrBadRequest
	}
	if actorID == otherID {
		return models.ErrInvalidSelfAction
	}
	// อาจเช็คก่อนก็ได้ว่าเป็นเพื่อนกันอยู่ไหม (optional)
	isFriend, err := s.friendsrepo.AreFriends(ctx, actorID, otherID)
	if err != nil {
		return err
	}
	if !isFriend {
		return models.ErrNotFriends
	}
	return s.friendsrepo.Unfriend(ctx, actorID, otherID)
}

/* 20-02 by ploy */

func (s *friendsService) SearchAddFriend(ctx context.Context, actorID int, search string, page, size int) ([]models.UserSearchItem, int, error) {
	if actorID == 0 {
		return nil, 0, ErrBadRequest
	}

	search = normalizeSearch(search)
	if search == "" {
		return []models.UserSearchItem{}, 0, nil
	}

	page, size = clampPageSize(page, size)
	limit, offset := toLimitOffset(page, size)

	items, err := s.friendsrepo.SearchAddFriend(ctx, actorID, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.friendsrepo.CountAddFriend(ctx, actorID, search)
	return items, total, err
}
