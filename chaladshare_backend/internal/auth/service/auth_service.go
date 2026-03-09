package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"chaladshare_backend/internal/auth/models"
	"chaladshare_backend/internal/auth/repository"
	"chaladshare_backend/internal/mail"
)

type AuthService interface {
	GetAllUsers() ([]models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	Register(email, username, password, verifyToken string) (*models.User, error)
	IsEmailTaken(email string) (bool, error)
	IsUsernameTaken(username string) (bool, error)
	Login(email, password string) (*models.User, error)
	IssueToken(userID int) (string, error)
	IssueSession(userID int) (accessToken string, refreshToken string, err error)
	ForgotPassword(email string) error
	ResetPassword(email, otp, newPassword string) error
	//88
	RequestEmailVerifyOTP(email string) error
	ConfirmEmailVerifyOTP(email, otp string) (string, error) // return verify_token
	ValidateEmailVerifyToken(email, token string) error
	VerifyForgotOTP(email, otp string) error

	LoginWithSession(email, password string) (accessToken string, refreshToken string, user *models.User, err error)
	Refresh(refreshToken string) (newAccess string, newRefresh string, err error)
	Logout(refreshToken string) error
}

type authService struct {
	userRepo        repository.AuthRepository
	jwtSecret       []byte
	tokenTTLMinutes int
	mailer          *mail.Mailer
}

func NewAuthService(userRepo repository.AuthRepository, secret []byte, ttlMin int) AuthService {
	// ถ้าไม่ได้ตั้งค่า SMTP ก็ให้ mailer เป็น nil (กันแอปล้มตอน dev)
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	var m *mail.Mailer

	p, err := strconv.Atoi(portStr)
	if err == nil && host != "" && user != "" && pass != "" && from != "" {
		m = mail.NewMailer(host, p, user, pass, from)
	}

	return &authService{
		userRepo:        userRepo,
		jwtSecret:       secret,
		tokenTTLMinutes: ttlMin,
		mailer:          m,
	}
}

func generateOTP6() (string, error) {
	// 000000 - 999999
	nBig, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", nBig.Int64()), nil
}

func (s *authService) IssueToken(userID int) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": userID,
		"iat":     now.Unix(),
		"exp":     now.Add(time.Duration(s.tokenTTLMinutes) * time.Minute).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(s.jwtSecret)
}

// ผู้ใช้ทั้งหมด
func (s *authService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.GetAllUsers()
}

func (s *authService) IssueSession(userID int) (string, string, error) {
	access, err := s.IssueToken(userID)
	if err != nil {
		return "", "", err
	}

	refresh, err := newRefreshToken()
	if err != nil {
		return "", "", err
	}

	refreshHash := hashToken(refresh)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	if _, err := s.userRepo.CreateSession(userID, refreshHash, expiresAt); err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

// ดึงผู้ใช้จากอีเมล
func (s *authService) GetUserByEmail(email string) (*models.User, error) {
	if strings.TrimSpace(email) == "" {
		return nil, errors.New("email is required")
	}
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// func register 88
func (s *authService) Register(email, username, password, verifyToken string) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	username = strings.TrimSpace(username)

	if email == "" || username == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("email, username and password are required")
	}

	// ✅ ตรวจ verify token ก่อนสมัคร (ต้องเรียก ValidateEmailVerifyToken)
	if err := s.ValidateEmailVerifyToken(email, verifyToken); err != nil {
		return nil, err
	}

	if !strings.Contains(email, "@") {
		return nil, errors.New("invalid email format")
	}
	taken, err := s.userRepo.IsEmailTaken(email)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, errors.New("email already in use")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	user, err := s.userRepo.CreateUser(email, username, string(hashedPassword))
	if err != nil {
		return nil, fmt.Errorf("cannot create user: %v", err)
	}

	return user, nil
}

// func login
func (s *authService) Login(email, password string) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" || strings.TrimSpace(password) == "" {
		return nil, errors.New("email and password are required")
	}

	// ดึงข้อมูลผู้ใช้จาก email
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil || user == nil {
		return nil, errors.New("invalid email")
	}

	// ตรวจสอบรหัสผ่าน
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}
func (s *authService) ForgotPassword(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return nil
	}

	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil || user == nil {
		return nil // กัน enumeration
	}

	otp, err := generateOTP6()
	if err != nil {
		return err
	}

	otpHash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(3 * time.Minute)

	// ปิด OTP เก่าที่ค้างอยู่
	_ = s.userRepo.MarkAllActivePasswordResetsUsed(user.ID)

	// ✅ บันทึก OTP ใหม่ (ห้ามลืม)
	if err := s.userRepo.CreatePasswordReset(user.ID, string(otpHash), expiresAt); err != nil {
		return err
	}

	subject := "ChaladShare OTP สำหรับรีเซ็ตรหัสผ่าน"
	body := fmt.Sprintf(
		"สวัสดี, คุณได้ทำการขอรีเซ็ตรหัสผ่าน รหัส OTP ของคุณคือ: %s\n\nกรุณาใช้งานภายใน 3 นาที\nหากคุณไม่ได้ทำรายการดังกล่าว กรุณาไม่ต้องดำเนินการใดๆ ",
		otp,
	)

	if s.mailer != nil {
		if err := s.mailer.Send(email, subject, body); err != nil {
			log.Println("send otp email failed:", err)
		}
	} else {
		log.Println("[FORGOT_PASSWORD] email:", email, "OTP:", otp)
	}

	return nil
}

func (s *authService) ResetPassword(email, otp, newPassword string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	otp = strings.TrimSpace(otp)
	newPassword = strings.TrimSpace(newPassword)

	if email == "" || otp == "" || newPassword == "" {
		return errors.New("missing fields")
	}

	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil || user == nil {
		return errors.New("invalid otp or expired")
	}

	pr, err := s.userRepo.GetLatestActivePasswordReset(user.ID)
	if err != nil {
		return errors.New("invalid otp or expired")
	}

	if time.Now().After(pr.ExpiresAt) {
		return errors.New("invalid otp or expired")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pr.OTPHash), []byte(otp)); err != nil {
		return errors.New("invalid otp or expired")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdateUserPasswordHash(user.ID, string(newHash)); err != nil {
		return err
	}

	if err := s.userRepo.MarkPasswordResetUsed(pr.ID); err != nil {
		return err
	}

	return nil
}

// 88
func (s *authService) VerifyForgotOTP(email, otp string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	otp = strings.TrimSpace(otp)

	if email == "" || otp == "" {
		return errors.New("missing fields")
	}

	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil || user == nil {
		return errors.New("invalid otp or expired")
	}

	pr, err := s.userRepo.GetLatestActivePasswordReset(user.ID)
	if err != nil {
		return errors.New("invalid otp or expired")
	}

	// ✅ กันไว้เพิ่ม (เผื่อ repo ยังไม่กรอง expires)
	if time.Now().After(pr.ExpiresAt) {
		return errors.New("invalid otp or expired")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(pr.OTPHash), []byte(otp)); err != nil {
		return errors.New("invalid otp or expired")
	}

	// ✅ สำคัญ: “ตรวจอย่างเดียว” ห้าม MarkUsed / ห้ามแก้รหัสผ่าน
	return nil
}
func (s *authService) RequestEmailVerifyOTP(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return nil // ไม่บอกอะไร (กัน abuse)
	}

	// ถ้ามี user อยู่แล้ว ก็ไม่ต้องส่ง OTP (กัน spam) แต่ยังตอบ OK แบบกลาง ๆ
	if existing, _ := s.userRepo.GetUserByEmail(email); existing != nil {
		return nil
	}

	otp, err := generateOTP6()
	if err != nil {
		return err
	}
	otpHash, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(3 * time.Minute)

	_ = s.userRepo.MarkAllActiveEmailVerificationsUsed(email)
	if err := s.userRepo.CreateEmailVerification(email, string(otpHash), expiresAt); err != nil {
		return err
	}

	subject := "ChaladShare OTP สำหรับยืนยันอีเมล"
	body := fmt.Sprintf(
		"สวัสดี, คุณได้ทำการขอสมัครสมาชิก รหัส OTP สำหรับยืนยันอีเมลของคุณคือ: %s\n\nกรุณาใช้งานภายใน 3 นาที\nหากคุณไม่ได้ทำรายการดังกล่าว กรุณาไม่ต้องดำเนินการใดๆ ",
		otp,
	)

	if s.mailer != nil {
		if err := s.mailer.Send(email, subject, body); err != nil {
			log.Println("send verify email otp failed:", err)
		}
	} else {
		log.Println("[VERIFY_EMAIL] email:", email, "OTP:", otp)
	}
	return nil
}

func (s *authService) ConfirmEmailVerifyOTP(email, otp string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	otp = strings.TrimSpace(otp)
	if email == "" || otp == "" {
		return "", errors.New("missing fields")
	}

	// ถ้ามี user อยู่แล้ว ไม่ให้ใช้ flow นี้
	if existing, _ := s.userRepo.GetUserByEmail(email); existing != nil {
		return "", errors.New("email already in use")
	}

	ev, err := s.userRepo.GetLatestActiveEmailVerification(email)
	if err != nil {
		return "", errors.New("invalid otp or expired")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(ev.OTPHash), []byte(otp)); err != nil {
		return "", errors.New("invalid otp or expired")
	}

	// ใช้แล้วปิด OTP
	_ = s.userRepo.MarkEmailVerificationUsed(ev.ID)

	// ✅ ออก verify_token (JWT อายุสั้น)
	now := time.Now()
	claims := jwt.MapClaims{
		"email":   email,
		"purpose": "verify_email",
		"iat":     now.Unix(),
		"exp":     now.Add(15 * time.Minute).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString(s.jwtSecret)
	if err != nil {
		return "", errors.New("issue verify token failed")
	}
	return token, nil
}

func (s *authService) ValidateEmailVerifyToken(email, token string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	token = strings.TrimSpace(token)
	if email == "" || token == "" {
		return errors.New("email verification required")
	}

	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token")
		}
		return s.jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		return errors.New("email verification required")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("email verification required")
	}
	purpose, _ := claims["purpose"].(string)
	em, _ := claims["email"].(string)

	if purpose != "verify_email" || strings.ToLower(em) != email {
		return errors.New("email verification required")
	}
	return nil
}
func (s *authService) IsEmailTaken(email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return false, errors.New("email is required")
	}
	return s.userRepo.IsEmailTaken(email)
}

func (s *authService) IsUsernameTaken(username string) (bool, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" {
		return false, errors.New("username is required")
	}
	return s.userRepo.IsUsernameTaken(username)
}

func (s *authService) LoginWithSession(email, password string) (accessToken string, refreshToken string, user *models.User, err error) {
	user, err = s.Login(email, password) // ใช้ของเดิม ไม่ต้องแก้
	if err != nil {
		return "", "", nil, err
	}

	accessToken, err = s.IssueToken(user.ID)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err = newRefreshToken()
	if err != nil {
		return "", "", nil, err
	}

	refreshHash := hashToken(refreshToken)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	if _, err := s.userRepo.CreateSession(user.ID, refreshHash, expiresAt); err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

func (s *authService) Refresh(refreshToken string) (newAccess string, newRefresh string, err error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return "", "", errors.New("missing refresh token")
	}

	hash := hashToken(refreshToken)
	sess, err := s.userRepo.GetSessionByRefresh(hash)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	if time.Now().After(sess.ExpiresAt) {
		_ = s.userRepo.RevokeSession(sess.SessionID)
		return "", "", errors.New("refresh token expired")
	}

	// ออก access token ใหม่
	newAccess, err = s.IssueToken(sess.UserID)
	if err != nil {
		return "", "", err
	}

	// rotate refresh token (ปลอดภัยกว่า)
	newRefresh, err = newRefreshToken()
	if err != nil {
		return "", "", err
	}
	newHash := hashToken(newRefresh)
	newExpires := time.Now().Add(30 * 24 * time.Hour)

	if _, err := s.userRepo.RotateSession(sess.SessionID, newHash, newExpires); err != nil {
		return "", "", err
	}

	return newAccess, newRefresh, nil
}

func (s *authService) Logout(refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil // idempotent
	}

	hash := hashToken(refreshToken)
	sess, err := s.userRepo.GetSessionByRefresh(hash)
	if err != nil {
		return nil // ไม่เจอก็ถือว่า logout แล้ว
	}
	return s.userRepo.RevokeSession(sess.SessionID)
}
