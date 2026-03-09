package service

import (
	"context"
	"fmt"
	"strings"

	fileservice "chaladshare_backend/internal/files/service"
	friendservice "chaladshare_backend/internal/friends/service"

	"chaladshare_backend/internal/posts/models"
	"chaladshare_backend/internal/posts/repository"
)

type PostService interface {
	CreatePost(post *models.Post, tags []string) (int, error)
	UpdatePost(post *models.Post, tags []string) error
	DeletePost(postID int) error

	GetAllPosts() ([]models.PostResponse, error)
	GetFeedPosts(viewerID int) ([]models.PostResponse, error)
	GetPostByID(postID int) (*models.PostResponse, error)
	GetPostByIDForViewer(viewerID, postID int) (*models.PostResponse, error)
	CountByUserID(userID int) (int, error)

	IsOwner(postID int, userID int) (bool, error)
	ViewPost(viewerID, postID int) (bool, string, error)
	Friends(viewerID, authorID int) (bool, error)

	GetSavedPosts(userID int) ([]models.PostResponse, error)
	GetPopularPosts(viewerID, limit int) ([]models.PostResponse, error)
	SearchPosts(viewerID int, search string, page, size int) ([]models.PostResponse, int, error)
}

type postService struct {
	postRepo  repository.PostRepository
	friendSvc friendservice.FriendService
	fileSvc   fileservice.FileService
}

func NewPostService(postRepo repository.PostRepository, friendSvc friendservice.FriendService, fileSvc fileservice.FileService) PostService {
	return &postService{postRepo: postRepo, friendSvc: friendSvc, fileSvc: fileSvc}
}

func normalizeVisibility(v string) (string, error) {
	vis := strings.ToLower(strings.TrimSpace(v))
	switch vis {
	case "", models.VisibilityPublic:
		return models.VisibilityPublic, nil
	case models.VisibilityFriends:
		return models.VisibilityFriends, nil
	default:
		return "", fmt.Errorf("unsupported visibility: %s", v)
	}
}

// สร้างโพสต์ใหม่
func (s *postService) CreatePost(post *models.Post, tags []string) (int, error) {
	if post.AuthorUserID <= 0 {
		return 0, fmt.Errorf("invalid author")
	}
	if strings.TrimSpace(post.Title) == "" {
		return 0, fmt.Errorf("post_title is required")
	}

	vis, err := normalizeVisibility(post.Visibility)
	if err != nil {
		return 0, err
	}
	post.Visibility = vis

	normTags := normalizeTags(tags)
	postID, err := s.postRepo.CreatePost(post, normTags)
	if err != nil {
		return 0, fmt.Errorf("failed to create post: %w", err)
	}
	return postID, nil
}

func (s *postService) UpdatePost(post *models.Post, tags []string) error {
	if post.PostID <= 0 {
		return fmt.Errorf("invalid post_id")
	}
	if strings.TrimSpace(post.Title) == "" {
		return fmt.Errorf("post_title is required")
	}

	visInput := strings.TrimSpace(post.Visibility)
	if visInput == "" {
		existing, err := s.postRepo.GetPostByID(post.PostID)
		if err != nil {
			return fmt.Errorf("get existing post: %w", err)
		}
		if existing == nil {
			return fmt.Errorf("post not found")
		}
		visInput = existing.Visibility
	}

	vis, err := normalizeVisibility(visInput)
	if err != nil {
		return err
	}
	post.Visibility = vis

	var normTags []string
	if tags != nil {
		normTags = normalizeTags(tags)
	}

	if err := s.postRepo.UpdatePost(post, normTags); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}
	return nil
}

func normalizeTags(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))

	const (
		maxTags = 10
		maxLen  = 30
	)
	for _, t := range in {
		tag := strings.TrimSpace(t)
		tag = strings.TrimPrefix(tag, "#")
		tag = strings.ToLower(tag)

		if tag == "" || len(tag) > maxLen {
			continue
		}
		valid := true
		for _, r := range tag {
			if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '_' && r != '-' {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}

		if _, dup := seen[tag]; dup {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
		if len(out) >= maxTags {
			break
		}
	}
	return out
}

func (s *postService) DeletePost(postID int) error {

	p, err := s.postRepo.GetPostByID(postID)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("post not found")
	}
	if err := s.postRepo.DeletePost(postID); err != nil {
		return err
	}
	if p.DocumentID != nil && s.fileSvc != nil {
		_ = s.fileSvc.DeleteFile(*p.DocumentID)
	}
	return nil
}

func (s *postService) GetAllPosts() ([]models.PostResponse, error) {
	return s.postRepo.GetAllPosts()
}

func (s *postService) GetFeedPosts(viewerID int) ([]models.PostResponse, error) {
	return s.postRepo.GetFeedPosts(viewerID)
}

// each post by ID
func (s *postService) GetPostByID(postID int) (*models.PostResponse, error) {
	return s.postRepo.GetPostByID(postID)
}

func (s *postService) GetPostByIDForViewer(viewerID, postID int) (*models.PostResponse, error) {
	return s.postRepo.GetPostByIDForViewer(viewerID, postID)
}

func (s *postService) CountByUserID(userID int) (int, error) {
	return s.postRepo.CountByUserID(userID)
}

func (s *postService) IsOwner(postID int, userID int) (bool, error) {
	ownerID, err := s.postRepo.GetPostOwnerID(postID)
	if err != nil {
		return false, fmt.Errorf("cannot get post owner: %w", err)
	}
	return ownerID == userID, nil
}

func (s *postService) ViewPost(viewerID, postID int) (bool, string, error) {
	post, err := s.GetPostByID(postID)
	if err != nil {
		return false, "error", fmt.Errorf("get post: %w", err)
	}
	if post == nil {
		return false, "not_found", nil
	}

	authorID := post.AuthorID
	vis := strings.ToLower(strings.TrimSpace(post.Visibility))
	if vis == "" {
		vis = models.VisibilityPublic
	}
	if viewerID == authorID {
		return true, "owner", nil
	}

	switch vis {
	case models.VisibilityPublic:
		return true, "public", nil
	case models.VisibilityFriends:
		ok, err := s.Friends(viewerID, authorID)
		if err != nil {
			return false, "error", err
		}
		if ok {
			return true, "friends", nil
		}
		return false, "friends_only", nil
	default:
		return false, "denied", nil
	}
}

func (s *postService) Friends(viewerID, authorID int) (bool, error) {
	if viewerID <= 0 || authorID <= 0 {
		return false, fmt.Errorf("invalid user id")
	}

	if viewerID == authorID {
		return true, nil
	}

	ok, err := s.friendSvc.AreFriends(context.Background(), viewerID, authorID)
	if err != nil {
		return false, fmt.Errorf("check friends: %w", err)
	}
	return ok, nil
}

func (s *postService) GetSavedPosts(userID int) ([]models.PostResponse, error) {
	return s.postRepo.GetSavedPosts(userID)
}

func (s *postService) GetPopularPosts(viewerID, limit int) ([]models.PostResponse, error) {
	if viewerID <= 0 {
		return nil, fmt.Errorf("invalid viewer id")
	}
	if limit <= 0 {
		limit = 3
	}
	if limit > 20 {
		limit = 20
	}
	return s.postRepo.GetPopularPosts(viewerID, limit)
}

func (s *postService) SearchPosts(viewerID int, search string, page, size int) ([]models.PostResponse, int, error) {
	if viewerID <= 0 {
		return nil, 0, fmt.Errorf("invalid viewer id")
	}

	search = strings.TrimSpace(search)

	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}
	return s.postRepo.SearchPosts(viewerID, search, page, size)
}
