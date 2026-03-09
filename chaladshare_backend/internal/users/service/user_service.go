// internal/users/service/service.go
package service

import (
	"context"
	"errors"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"chaladshare_backend/internal/users/models"
	"chaladshare_backend/internal/users/repository"
)

type UserService interface {
	GetOwnProfile(ctx context.Context, userID int) (*models.OwnProfileResponse, error)
	GetViewedUserProfile(ctx context.Context, userID int) (*models.ViewedUserProfileResponse, error)
	UpdateOwnProfile(ctx context.Context, userID int, req *models.UpdateOwnProfileRequest) error
	ChangePassword(ctx context.Context, userID int, currentPwd, newPwd string) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{repo: r}
}

func (s *userService) GetOwnProfile(ctx context.Context, userID int) (*models.OwnProfileResponse, error) {
	return s.repo.GetOwnProfile(ctx, userID)
}

func (s *userService) GetViewedUserProfile(ctx context.Context, userID int) (*models.ViewedUserProfileResponse, error) {
	return s.repo.GetViewedUserProfile(ctx, userID)
}

func (s *userService) UpdateOwnProfile(ctx context.Context, userID int, req *models.UpdateOwnProfileRequest) error {
	if req == nil || (req.Username == nil && req.AvatarURL == nil && req.AvatarStore == nil && req.Bio == nil) {
		return errors.New("no fields to update")
	}

	if req.Username != nil {
		l := utf8.RuneCountInString(*req.Username)
		if l < 3 || l > 50 {
			return errors.New("username must be 3–50 characters")
		}
	}
	if req.Bio != nil {
		if utf8.RuneCountInString(*req.Bio) > 150 {
			return errors.New("bio must be at most 150 characters")
		}
	}
	return s.repo.UpdateOwnProfile(ctx, userID, req)
}

func (s *userService) ChangePassword(ctx context.Context, userID int, current string, newPwd string) error {
	if len(current) == 0 || len(newPwd) == 0 {
		return errors.New("กรุณากรอกรหัสผ่านให้ครบ")
	}
	if utf8.RuneCountInString(newPwd) < 8 {
		return errors.New("รหัสผ่านใหม่ต้องมีอย่างน้อย 8 ตัวอักษร")
	}

	oldHash, err := s.repo.GetPasswordHash(ctx, userID)
	if err != nil {
		return errors.New("ไม่พบผู้ใช้")
	}

	// ตรวจรหัสเดิม
	if err := bcrypt.CompareHashAndPassword([]byte(oldHash), []byte(current)); err != nil {
		return errors.New("รหัสผ่านปัจจุบันไม่ถูกต้อง")
	}

	// สร้าง hash ใหม่
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("ไม่สามารถตั้งรหัสผ่านใหม่ได้")
	}

	if err := s.repo.UpdatePasswordHash(ctx, userID, string(newHash)); err != nil {
		return errors.New("อัปเดตรหัสผ่านไม่สำเร็จ")
	}
	return nil
}
