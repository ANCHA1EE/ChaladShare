package service

import "chaladshare_backend/internal/posts/repository"

type SaveService interface {
	ToggleSave(userID, postID int) (isSaved bool, saveCount int, err error)
	IsPostSaved(userID, postID int) (bool, error)
}

type saveService struct {
	saveRepo repository.SaveRepository
}

func NewSaveService(saveRepo repository.SaveRepository) SaveService {
	return &saveService{saveRepo: saveRepo}
}

func (s *saveService) ToggleSave(userID, postID int) (bool, int, error) {
	// 1) เช็กก่อนว่า user นี้เคยบันทึกโพสต์นี้หรือยัง
	saved, err := s.saveRepo.IsPostSaved(userID, postID)
	if err != nil {
		return false, 0, err
	}

	// 2) ถ้าเคย save แล้ว → กดอีกที = unsave
	if saved {
		if err := s.saveRepo.UnsavePost(userID, postID); err != nil {
			return false, 0, err
		}
		saved = false
	} else {
		// ถ้ายังไม่เคย save → save ใหม่
		if err := s.saveRepo.SavePost(userID, postID); err != nil {
			return false, 0, err
		}
		saved = true
	}

	// 3) ดึงจำนวนบันทึกล่าสุด
	count, err := s.saveRepo.SaveCount(postID)
	if err != nil {
		return false, 0, err
	}

	return saved, count, nil
}

//ตรวจสอบ
func (s *saveService) IsPostSaved(userID, postID int) (bool, error) {
	return s.saveRepo.IsPostSaved(userID, postID)
}
