// internal/users/handlers/handler.go
package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	friendservice "chaladshare_backend/internal/friends/service"
	postsvc "chaladshare_backend/internal/posts/service"
	"chaladshare_backend/internal/users/models"
	"chaladshare_backend/internal/users/service"
)

type UserHandler struct {
	userSvc    service.UserService
	postSvc    postsvc.PostService
	friendsSvc friendservice.FriendService
}

func NewUserHandler(s service.UserService, p postsvc.PostService, f friendservice.FriendService) *UserHandler {
	return &UserHandler{userSvc: s, postSvc: p, friendsSvc: f}
}

func getUID(c *gin.Context) (int, bool) {
	for _, k := range []string{"user_id", "uid"} {
		if v, ok := c.Get(k); ok {
			switch t := v.(type) {
			case int:
				return t, true
			case int64:
				return int(t), true
			case float64:
				return int(t), true
			}
		}
	}
	return 0, false
}

func (h *UserHandler) GetOwnProfile(c *gin.Context) {
	uid, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	prof, err := h.userSvc.GetOwnProfile(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	withSet := map[string]bool{}
	if raw := strings.TrimSpace(c.Query("with")); raw != "" {
		for _, w := range strings.Split(raw, ",") {
			withSet[strings.ToLower(strings.TrimSpace(w))] = true
		}
	}
	want := func(k string) bool { return withSet["all"] || withSet[k] }

	// base response เดิม
	resp := gin.H{
		"user_id":         prof.UserID,
		"email":           prof.Email,
		"username":        prof.Username,
		"avatar_url":      prof.AvatarURL,
		"avatar_storage":  prof.AvatarStore,
		"bio":             prof.Bio,
		"user_status":     prof.Status,
		"user_created_at": prof.CreatedAt,
	}

	if want("stats") {
		if cnt, err := h.postSvc.CountByUserID(uid); err == nil {
			resp["posts_count"] = cnt
		} else {
			resp["posts_count"] = 0
		}
	}
	if want("followers") || want("following") {
		fols, folg, _, err := h.friendsSvc.GetFollowStats(c.Request.Context(), uid)
		if err == nil {
			if want("followers") {
				resp["followers_count"] = fols
			}
			if want("following") {
				resp["following_count"] = folg
			}
		} else {
			if want("followers") {
				resp["followers_count"] = 0
			}
			if want("following") {
				resp["following_count"] = 0
			}
		}
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) GetViewedUserProfile(c *gin.Context) {
	viewerID, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil || targetID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	prof, err := h.userSvc.GetViewedUserProfile(c.Request.Context(), targetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	withSet := map[string]bool{}
	if raw := strings.TrimSpace(c.Query("with")); raw != "" {
		for _, w := range strings.Split(raw, ",") {
			withSet[strings.ToLower(strings.TrimSpace(w))] = true
		}
	}
	want := func(k string) bool { return withSet["all"] || withSet[k] }

	isFollowing := false
	if viewerID != targetID {
		if ok2, err := h.friendsSvc.IsFollowing(c.Request.Context(), viewerID, targetID); err == nil {
			isFollowing = ok2
		}
	}

	resp := gin.H{
		"user_id":        prof.UserID,
		"username":       prof.Username,
		"avatar_url":     prof.AvatarURL,
		"avatar_storage": prof.AvatarStore,
		"bio":            prof.Bio,
		"is_following":   isFollowing,
	}

	if want("stats") {
		if cnt, err := h.postSvc.CountByUserID(targetID); err == nil {
			resp["posts_count"] = cnt
		} else {
			resp["posts_count"] = 0
		}
	}
	if want("followers") || want("following") {
		fols, folg, _, err := h.friendsSvc.GetFollowStats(c.Request.Context(), targetID)
		if err == nil {
			if want("followers") {
				resp["followers_count"] = fols
			}
			if want("following") {
				resp["following_count"] = folg
			}
		} else {
			if want("followers") {
				resp["followers_count"] = 0
			}
			if want("following") {
				resp["following_count"] = 0
			}
		}
	}
	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) UpdateOwnProfile(c *gin.Context) {
	uid, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req models.UpdateOwnProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	if err := h.userSvc.UpdateOwnProfile(c.Request.Context(), uid, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ===== Change Password =====

type changePasswordReq struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	uid, ok := getUID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ยืนยันรหัสผ่านใหม่ไม่ตรงกัน"})
		return
	}

	if err := h.userSvc.ChangePassword(c.Request.Context(), uid, req.CurrentPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
