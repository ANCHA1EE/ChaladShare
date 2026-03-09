package models

import (
	"time"
)

const (
	VisibilityPublic  = "public"
	VisibilityFriends = "friends"
)

// post
type Post struct {
	PostID       int       `json:"post_id"`
	AuthorUserID int       `json:"-"`
	Title        string    `json:"post_title"`
	Description  string    `json:"post_description"`
	Visibility   string    `json:"post_visibility"`
	DocumentID   *int      `json:"post_document_id"`
	CoverURL     *string   `json:"post_cover_url"`
	CreatedAt    time.Time `json:"post_created_at"`
	UpdatedAt    time.Time `json:"post_updated_at"`
}

// each tag
type Tag struct {
	TagID   int    `json:"tag_id"`
	TagName string `json:"tag_name"`
}

type PostTag struct {
	PostID int `json:"post_id"`
	TagID  int `json:"tag_id"`
}

// liked
type Like struct {
	UserID    int       `json:"like_user_id"`
	PostID    int       `json:"like_post_id"`
	CreatedAt time.Time `json:"like_created_at"`
}

// save post
type SavePost struct {
	UserID    int       `json:"save_user_id"`
	PostID    int       `json:"save_post_id"`
	CreatedAt time.Time `json:"save_created_at"`
}

// post stat
type PostStats struct {
	PostID         int       `json:"post_id"`
	LikeCount      int       `json:"like_count"`
	SaveCount      int       `json:"save_count"`
	LastActivityAt time.Time `json:"last_activity_at"`
}

// for response
type PostResponse struct {
	PostID       int       `json:"post_id"`
	AuthorID     int       `json:"author_id"`
	AuthorName   string    `json:"author_name"`
	Title        string    `json:"post_title"`
	Description  string    `json:"post_description"`
	Visibility   string    `json:"post_visibility"`
	DocumentID   *int      `json:"post_document_id"`
	DocumentName *string   `json:"document_name"`
	CreatedAt    time.Time `json:"post_created_at"`
	UpdatedAt    time.Time `json:"post_updated_at"`

	FileURL   *string  `json:"file_url"`
	CoverURL  *string  `json:"cover_url"`
	AvatarURL *string  `json:"avatar_url"`
	Tags      []string `json:"tags"`
	LikeCount int      `json:"like_count"`
	SaveCount int      `json:"save_count"`

	IsLiked bool `json:"is_liked"`
	IsSaved bool `json:"is_saved"`
}

type UpdatePostRequest struct {
	Title       string   `json:"post_title,omitempty"`
	Description string   `json:"post_description,omitempty"`
	Visibility  string   `json:"post_visibility,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type PostQueryParam struct {
	Search string   `form:"search"`
	Tag    []string `form:"tag"`
	Sort   string   `form:"sort"`
	Limit  int      `form:"limit"`
}
