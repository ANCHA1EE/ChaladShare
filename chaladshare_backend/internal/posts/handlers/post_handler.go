package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"chaladshare_backend/internal/posts/models"
	"chaladshare_backend/internal/posts/service"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	postService service.PostService
	saveService service.SaveService
}

func NewPostHandler(postService service.PostService, saveService service.SaveService) *PostHandler {
	return &PostHandler{
		postService: postService,
		saveService: saveService,
	}
}

// สร้างโพสต์ใหม่ (ต้องล็อกอิน)
func (h *PostHandler) CreatePost(c *gin.Context) {
	uid := c.GetInt("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		Title       string   `json:"post_title" binding:"required"`
		Description string   `json:"post_description"`
		Visibility  string   `json:"post_visibility" binding:"required"` // ตอนนี้รองรับ "public" เท่านั้น
		DocumentID  *int     `json:"document_id"`
		CoverURL    *string  `json:"cover_url"`
		Tags        []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if req.Visibility != models.VisibilityPublic && req.Visibility != models.VisibilityFriends {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported visibility"})
		return
	}

	post := &models.Post{
		AuthorUserID: uid,
		Title:        req.Title,
		Description:  req.Description,
		Visibility:   req.Visibility,
		DocumentID:   req.DocumentID,
		CoverURL:     req.CoverURL,
	}

	postID, err := h.postService.CreatePost(post, req.Tags)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Location", "/api/v1/posts/"+strconv.Itoa(postID))
	c.JSON(http.StatusCreated, gin.H{"data": gin.H{"post_id": postID}})
}

// ดึงโพสต์ทั้งหมด (ต้องล็อกอิน)
func (h *PostHandler) GetAllPosts(c *gin.Context) {
	uid := c.GetInt("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	posts, err := h.postService.GetFeedPosts(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": posts})
}

// รายละเอียดโพสต์ (ต้องล็อกอิน)
func (h *PostHandler) GetPostByID(c *gin.Context) {
	uid := c.GetInt("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ok, reason, err := h.postService.ViewPost(uid, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		switch reason {
		case "not_found":
			c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		case "friends_only", "denied":
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		}
		return
	}

	post, err := h.postService.GetPostByIDForViewer(uid, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if post == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "post not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": post})
}

// แก้ไขโพสต์ (เฉพาะเจ้าของ)
func (h *PostHandler) UpdatePost(c *gin.Context) {
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

	isOwner, err := h.postService.IsOwner(postID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	var req struct {
		Title       string   `json:"post_title"`
		Description string   `json:"post_description"`
		Visibility  *string  `json:"post_visibility"`
		Tags        []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	vis := ""
	if req.Visibility != nil {
		v := strings.ToLower(strings.TrimSpace(*req.Visibility))
		if v != models.VisibilityPublic && v != models.VisibilityFriends && v != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported visibility"})
			return
		}
		vis = v
	}

	post := &models.Post{
		PostID:      postID,
		Title:       req.Title,
		Description: req.Description,
		Visibility:  vis,
	}
	if err := h.postService.UpdatePost(post, req.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post updated successfully"})
}

// ลบโพสต์ (เฉพาะเจ้าของ)
func (h *PostHandler) DeletePost(c *gin.Context) {
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
	// เช็คสิทธิ์เจ้าของก่อน
	isOwner, err := h.postService.IsOwner(postID, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isOwner {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if err := h.postService.DeletePost(postID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ดึงรายการโพสต์ที่ user คนนี้บันทึกไว้
func (h *PostHandler) GetSavedPosts(c *gin.Context) {
	uid := c.GetInt("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	posts, err := h.postService.GetSavedPosts(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": posts})
}

// toggle save
func (h *PostHandler) ToggleSave(c *gin.Context) {
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

	isSaved, saveCount, err := h.saveService.ToggleSave(uid, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"post_id":    postID,
			"is_saved":   isSaved,
			"save_count": saveCount,
		},
	})
}

func (h *PostHandler) GetPopularPosts(c *gin.Context) {
	uid := c.GetInt("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limitStr := c.DefaultQuery("limit", "3")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	posts, err := h.postService.GetPopularPosts(uid, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": posts})
}

func (h *PostHandler) SearchPosts(c *gin.Context) {
	uid := c.GetInt("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	search := strings.TrimSpace(c.Query("search"))

	if search == "" {
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"items": []models.PostResponse{},
				"total": 0,
				"page":  1,
				"size":  20,
			},
		})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 || size > 100 {
		size = 20
	}

	items, total, err := h.postService.SearchPosts(uid, search, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"items":  items,
			"total":  total,
			"page":   page,
			"size":   size,
			"search": search,
		},
	})
}
