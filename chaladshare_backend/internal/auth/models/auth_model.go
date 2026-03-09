package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	Status       string    `json:"status"`
}

// register
type RegisterRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	VerifyToken string `json:"verify_token"` // ✅ ต้องมีเพื่อสมัครได้ 88
}

// login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// response ส่งกลับให้ client
type AuthResponse struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	// Token 	  string 	`json:"token,omitempty"`
}
type PasswordReset struct {
	ID        int
	UserID    int
	OTPHash   string
	ExpiresAt time.Time
	UsedAt    *time.Time
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

// ✅ ขอ OTP เพื่อยืนยันอีเมล 88
type RequestEmailVerifyOTPRequest struct {
	Email string `json:"email"`
}

// ✅ ยืนยัน OTP แล้วรับ verify_token 88
type ConfirmEmailVerifyOTPRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type ConfirmEmailVerifyOTPResponse struct {
	VerifyToken string `json:"verify_token"`
}
type EmailVerification struct {
	ID        int
	Email     string
	OTPHash   string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

type RefreshResponse struct {
	Message string `json:"message"`
}

type AuthSession struct {
	SessionID        int
	UserID           int
	RefreshTokenHash string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	ReplacedByID     *int
	CreatedAt        time.Time
	LastUsedAt       *time.Time
}
