package models

type User struct {
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Status       string `json:"user_status"`
	CreatedAt    string `json:"user_created_at"`
}

type Profile struct {
	ProfileUserID int     `json:"profile_user_id"`
	AvatarURL     *string `json:"avatar_url"`
	AvatarStore   *string `json:"avatar_storage"`
	Bio           *string `json:"bio"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type UserProfile struct {
	UserID      int64   `db:"user_id"`
	Email       string  `db:"email"`
	Username    string  `db:"username"`
	Status      string  `db:"user_status"`
	CreatedAt   string  `db:"user_created_at"`
	AvatarURL   *string `db:"avatar_url"`
	AvatarStore *string `db:"avatar_storage"`
	Bio         *string `db:"bio"`
}

// change password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func (j UserProfile) ToOwnProfileResponse() OwnProfileResponse {
	r := OwnProfileResponse{
		UserID:    j.UserID,
		Email:     j.Email,
		Username:  j.Username,
		Status:    j.Status,
		CreatedAt: j.CreatedAt,
	}
	if j.AvatarURL != nil {
		r.AvatarURL = *j.AvatarURL
	}
	if j.AvatarStore != nil {
		r.AvatarStore = *j.AvatarStore
	}
	if j.Bio != nil {
		r.Bio = *j.Bio
	}
	return r
}

// ดูของผู้ใช้คนอื่น
func (j UserProfile) ToViewedUserProfileResponse() ViewedUserProfileResponse {
	r := ViewedUserProfileResponse{
		UserID:   j.UserID,
		Username: j.Username,
	}
	if j.AvatarURL != nil {
		r.AvatarURL = *j.AvatarURL
	}
	if j.AvatarStore != nil {
		r.AvatarStore = *j.AvatarStore
	}
	if j.Bio != nil {
		r.Bio = *j.Bio
	}
	return r
}

type UpdateOwnProfileRequest struct {
	Username    *string `json:"username"`
	AvatarURL   *string `json:"avatar_url"`
	AvatarStore *string `json:"avatar_storage"`
	Bio         *string `json:"bio"`
}

type OwnProfileResponse struct {
	UserID      int64  `json:"user_id"`
	Email       string `json:"email"`
	Username    string `json:"username"`
	AvatarURL   string `json:"avatar_url"`
	AvatarStore string `json:"avatar_storage"`
	Bio         string `json:"bio"`
	Status      string `json:"user_status"`
	CreatedAt   string `json:"user_created_at"`
}

type ViewedUserProfileResponse struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	AvatarURL   string `json:"avatar_url"`
	AvatarStore string `json:"avatar_storage"`
	Bio         string `json:"bio"`
}
