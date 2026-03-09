package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"chaladshare_backend/internal/friends/models"
	"chaladshare_backend/internal/friends/service"
	"chaladshare_backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type FriendHandler struct {
	friendservice service.FriendService
}

func NewFriendHandler(friendservice service.FriendService) *FriendHandler {
	return &FriendHandler{friendservice: friendservice}
}

func getUID(c *gin.Context) (int, bool) {
	v, ok := c.Get(middleware.CtxUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok && id > 0
}

func parseParamID(c *gin.Context, key string) (int, bool) {
	raw := c.Param(key)
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + key})
		return 0, false
	}
	return n, true
}

func parsePageSize(c *gin.Context) (page, size int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ = strconv.Atoi(c.DefaultQuery("size", "20"))
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}
	return
}

func respondError(c *gin.Context, err error) {
	switch err {
	case service.ErrBadRequest:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case service.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case models.ErrInvalidSelfAction, models.ErrAlreadyFriends, models.ErrNotFriends:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *FriendHandler) FollowUser(c *gin.Context) {
	actorID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req models.CreateFollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := req.Validate(actorID); err != nil {
		respondError(c, err)
		return
	}

	if err := h.friendservice.FollowUser(c.Request.Context(), actorID, req.FollowedUserID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *FriendHandler) UnfollowUser(c *gin.Context) {
	actorID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	targetID, ok := parseParamID(c, "id")
	if !ok {
		return
	}

	if err := h.friendservice.UnfollowUser(c.Request.Context(), actorID, targetID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *FriendHandler) ListFriends(c *gin.Context) {
	viewerID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := parseParamID(c, "id")
	if !ok {
		return
	}
	search := c.DefaultQuery("search", "")
	page, size := parsePageSize(c)

	items, total, err := h.friendservice.ListFriends(c.Request.Context(), viewerID, userID, search, page, size)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": items, "total": total, "page": page, "size": size,
	})
}

/* 20-02 by ploy */

func (h *FriendHandler) SearchAddFriend(c *gin.Context) {
	actorID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	search := strings.TrimSpace(c.DefaultQuery("search", ""))
	page, size := parsePageSize(c)

	items, total, err := h.friendservice.SearchAddFriend(
		c.Request.Context(),
		actorID,
		search,
		page,
		size,
	)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items, "total": total, "page": page, "size": size,
	})
}

/* 20-02 by ploy */

func (h *FriendHandler) ListFollowers(c *gin.Context) {
	viewerID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := parseParamID(c, "id")
	if !ok {
		return
	}
	search := c.DefaultQuery("search", "")
	page, size := parsePageSize(c)

	items, total, err := h.friendservice.ListFollowers(c.Request.Context(), viewerID, userID, search, page, size)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": items, "total": total, "page": page, "size": size,
	})
}

func (h *FriendHandler) ListFollowing(c *gin.Context) {
	viewerID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := parseParamID(c, "id")
	if !ok {
		return
	}
	search := c.DefaultQuery("search", "")
	page, size := parsePageSize(c)

	items, total, err := h.friendservice.ListFollowing(c.Request.Context(), viewerID, userID, search, page, size)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": items, "total": total, "page": page, "size": size,
	})
}

func (h *FriendHandler) GetStats(c *gin.Context) {
	userID, ok := parseParamID(c, "id")
	if !ok {
		return
	}
	followers, following, friends, err := h.friendservice.GetFollowStats(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"followers": followers,
		"following": following,
		"friends":   friends,
	})
}

// Friend Requests

func (h *FriendHandler) SendFriendRequest(c *gin.Context) {
	actorID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req models.SendFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := req.Validate(actorID); err != nil {
		respondError(c, err)
		return
	}

	requestID, err := h.friendservice.SendFriendRequest(c.Request.Context(), actorID, req.ToUserID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"request_id": requestID})
}

func (h *FriendHandler) ListIncomingRequests(c *gin.Context) {
	actorID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	page, size := parsePageSize(c)

	items, total, err := h.friendservice.ListIncomingRequests(c.Request.Context(), actorID, page, size)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "size": size})
}

func (h *FriendHandler) ListOutgoingRequests(c *gin.Context) {
	actorID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	page, size := parsePageSize(c)

	items, total, err := h.friendservice.ListOutgoingRequests(c.Request.Context(), actorID, page, size)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "size": size})
}

func (h *FriendHandler) AcceptFriendRequest(c *gin.Context) {
	actorID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	reqID, ok := parseParamID(c, "id")
	if !ok {
		return
	}

	if err := h.friendservice.AcceptFriendRequest(c.Request.Context(), actorID, reqID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *FriendHandler) DeclineFriendRequest(c *gin.Context) {
	actorID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	reqID, ok := parseParamID(c, "id")
	if !ok {
		return
	}

	if err := h.friendservice.DeclineFriendRequest(c.Request.Context(), actorID, reqID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *FriendHandler) CancelFriendRequest(c *gin.Context) {
	actorID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	reqID, ok := parseParamID(c, "id")
	if !ok {
		return
	}

	if err := h.friendservice.CancelFriendRequest(c.Request.Context(), actorID, reqID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *FriendHandler) Unfriend(c *gin.Context) {
	actorID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	otherID, ok := parseParamID(c, "id")
	if !ok {
		return
	}

	if err := h.friendservice.Unfriend(c.Request.Context(), actorID, otherID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
