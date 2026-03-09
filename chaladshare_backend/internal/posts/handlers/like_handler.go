package handlers

import (
	"net/http"
	"strconv"

	"chaladshare_backend/internal/posts/service"

	"github.com/gin-gonic/gin"
)

type RecommendHook interface {
	OnLikeHook(userID int)
}

type LikeHandler struct {
	likeService      service.LikeService
	recommendService RecommendHook
}

func NewLikeHandler(likeService service.LikeService, recommendService RecommendHook) *LikeHandler {
	return &LikeHandler{
		likeService:      likeService,
		recommendService: recommendService,
	}
}

func (h *LikeHandler) ToggleLike(c *gin.Context) {
	uid := c.GetInt("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil || postID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	isLiked, likeCount, err := h.likeService.ToggleLike(uid, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// เรียกทุกครั้งหลัง toggle สำเร็จ ทั้ง like และ unlike
	if h.recommendService != nil {
		h.recommendService.OnLikeHook(uid)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"post_id":    postID,
			"is_liked":   isLiked,
			"like_count": likeCount,
		},
	})
}
