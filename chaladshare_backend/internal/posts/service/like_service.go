package service

import "chaladshare_backend/internal/posts/repository"

type LikeService interface {
	ToggleLike(userID, postID int) (isLiked bool, likeCount int, err error)
	IsPostLiked(userID, postID int) (bool, error)
}

type likeService struct {
	likeRepo repository.LikeRepository
}

func NewLikeService(likeRepo repository.LikeRepository) LikeService {
	return &likeService{likeRepo: likeRepo}
}

func (s *likeService) ToggleLike(userID, postID int) (bool, int, error) {
	// 1) เช็กก่อนว่าเคยไลก์หรือยัง
	liked, err := s.likeRepo.IsPostLiked(userID, postID)
	if err != nil {
		return false, 0, err
	}

	// 2) ถ้าเคยไลก์ → ยกเลิก / ถ้ายัง → ไลก์
	if liked {
		if err := s.likeRepo.UnlikePost(userID, postID); err != nil {
			return false, 0, err
		}
		liked = false
	} else {
		if err := s.likeRepo.LikePost(userID, postID); err != nil {
			return false, 0, err
		}
		liked = true
	}

	// 3) ดึงจำนวนไลก์ล่าสุด
	count, err := s.likeRepo.LikeCount(postID)
	if err != nil {
		return false, 0, err
	}

	return liked, count, nil
}

// ตรวจสอบ
func (s *likeService) IsPostLiked(userID, postID int) (bool, error) {
	return s.likeRepo.IsPostLiked(userID, postID)
}
