package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"chaladshare_backend/internal/auth/models"
	"chaladshare_backend/internal/auth/service"
)

type AuthHandler struct {
	authService       service.AuthService
	accessCookieName  string
	refreshCookieName string
	secure            bool
}

func NewAuthHandler(authService service.AuthService, accessCookieName string, refreshCookieName string, secure bool) *AuthHandler {
	return &AuthHandler{
		authService:       authService,
		accessCookieName:  accessCookieName,
		refreshCookieName: refreshCookieName,
		secure:            secure,
	}
}

// // ✅ สำคัญ: ข้ามโดเมน (Vercel) ต้อง SameSite=None และ Secure=true (ตอน prod)
// func (h *AuthHandler) setAuthCookie(c *gin.Context, token string) {
// 	http.SetCookie(c.Writer, &http.Cookie{
// 		Name:     h.accessCookieName,
// 		Value:    token,
// 		Path:     "/",
// 		HttpOnly: true,
// 		Secure:   h.secure,
// 		SameSite: http.SameSiteNoneMode,
// 	})
// }

// func (h *AuthHandler) clearAuthCookie(c *gin.Context) {
// 	http.SetCookie(c.Writer, &http.Cookie{
// 		Name:     h.accessCookieName,
// 		Value:    "",
// 		Path:     "/",
// 		MaxAge:   -1,
// 		HttpOnly: true,
// 		Secure:   h.secure,
// 		SameSite: http.SameSiteNoneMode,
// 	})
// }

// 88
func (h *AuthHandler) setAccessCookie(c *gin.Context, token string) {
	sameSite := http.SameSiteLaxMode
	if h.secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.accessCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: sameSite,
		MaxAge:   15 * 60,
	})
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, refreshToken string) {
	sameSite := http.SameSiteLaxMode
	if h.secure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.refreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: sameSite,
		MaxAge:   7 * 24 * 60 * 60,
	})
}

func (h *AuthHandler) clearAuthCookies(c *gin.Context) {
	sameSite := http.SameSiteLaxMode
	if h.secure {
		sameSite = http.SameSiteNoneMode
	}

	// clear access
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.accessCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: sameSite,
	})

	// clear refresh
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.refreshCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: sameSite,
	})
}

// Get all user
func (h *AuthHandler) GetAllUsers(c *gin.Context) {
	users, err := h.authService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// Register
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON format"})
		return
	}

	user, err := h.authService.Register(req.Email, req.Username, req.Password, req.VerifyToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// สร้าง access+refresh พร้อม session
	access, refresh, err := h.authService.IssueSession(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "issue session failed"})
		return
	}
	h.setAccessCookie(c, access)
	h.setRefreshCookie(c, refresh)

	resp := models.AuthResponse{
		ID: user.ID, Email: user.Email, Username: user.Username,
		CreatedAt: user.CreatedAt, Status: user.Status,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    resp,
	})
}

// Login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON format"})
		return
	}

	access, refresh, user, err := h.authService.LoginWithSession(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	h.setAccessCookie(c, access)
	h.setRefreshCookie(c, refresh)

	resp := models.AuthResponse{
		ID: user.ID, Email: user.Email, Username: user.Username,
		CreatedAt: user.CreatedAt, Status: user.Status,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"user":    resp,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// revoke session ใน DB (ถ้ามี refresh)
	if rt, err := c.Cookie(h.refreshCookieName); err == nil && strings.TrimSpace(rt) != "" {
		_ = h.authService.Logout(rt) // idempotent
	}

	// clear ทั้ง access + refresh
	h.clearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	rt, err := c.Cookie(h.refreshCookieName)
	if err != nil || strings.TrimSpace(rt) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	newAccess, newRefresh, err := h.authService.Refresh(rt)
	if err != nil {
		h.clearAuthCookies(c)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	h.setAccessCookie(c, newAccess)
	h.setRefreshCookie(c, newRefresh)

	c.JSON(http.StatusOK, models.RefreshResponse{
		Message: "refreshed",
	})
}

// ForgotPassword - ขอ OTP เพื่อรีเซ็ตรหัสผ่าน
// ForgotPassword - ขอ OTP เพื่อรีเซ็ตรหัสผ่าน
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	// ✅ ใช้ของเดิมที่มีอยู่แล้ว
	exists, err := h.authService.IsEmailTaken(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่มีอีเมลนี้ในระบบ"})
		return
	}

	// ถ้า ForgotPassword คืน error ได้ ใช้แบบนี้
	_ = h.authService.ForgotPassword(email)
	c.JSON(http.StatusOK, gin.H{"message": "ส่งรหัส OTP แล้ว กรุณาตรวจสอบอีเมล"})
}

// ResetPassword - ตรวจ OTP และตั้งรหัสผ่านใหม่
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.OTP) == "" || strings.TrimSpace(req.NewPassword) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
		return
	}

	if err := h.authService.ResetPassword(req.Email, req.OTP, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "reset password success",
	})
}

// ขอ OTP ยืนยันอีเมล 88
func (h *AuthHandler) RequestVerifyEmailOTP(c *gin.Context) {
	var req models.RequestEmailVerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	_ = h.authService.RequestEmailVerifyOTP(req.Email)

	// กัน enumeration: ตอบกลาง ๆ
	c.JSON(http.StatusOK, gin.H{"message": "ถ้าอีเมลนี้ใช้งานได้ ระบบจะส่ง OTP ให้"})
}

// ยืนยัน OTP แล้วรับ verify_token
func (h *AuthHandler) ConfirmVerifyEmailOTP(c *gin.Context) {
	var req models.ConfirmEmailVerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.OTP) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
		return
	}

	token, err := h.authService.ConfirmEmailVerifyOTP(req.Email, req.OTP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ConfirmEmailVerifyOTPResponse{VerifyToken: token})
}

// ขอ OTP สำหรับสมัครสมาชิก (ต้องบอกชัดเจนว่า email/username ซ้ำได้)
// POST /auth/register/request-otp
func (h *AuthHandler) RequestRegisterOTP(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON format"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	username := strings.TrimSpace(req.Username)

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณากรอกอีเมล"})
		return
	}
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณากรอกชื่อผู้ใช้"})
		return
	}

	// ✅ เช็ค email ซ้ำ
	emailTaken, err := h.authService.IsEmailTaken(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	if emailTaken {
		c.JSON(http.StatusConflict, gin.H{"error": "อีเมลนี้เคยสมัครไปแล้ว"})
		return
	}

	// ✅ เช็ค username ซ้ำ (ควรเช็คแบบ case-insensitive ให้ตรงกับ username_ci)
	usernameTaken, err := h.authService.IsUsernameTaken(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	if usernameTaken {
		c.JSON(http.StatusConflict, gin.H{"error": "ชื่อผู้ใช้นี้มีคนใช้แล้ว"})
		return
	}

	// ✅ ผ่านแล้วค่อยส่ง OTP (ใช้ flow เดิมของ verify email otp ได้)
	if err := h.authService.RequestEmailVerifyOTP(email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ส่ง OTP ไม่สำเร็จ"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ส่ง OTP แล้ว"})
}
func (h *AuthHandler) VerifyForgotPasswordOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.OTP) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
		return
	}

	if err := h.authService.VerifyForgotOTP(req.Email, req.OTP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "otp valid"})
}
