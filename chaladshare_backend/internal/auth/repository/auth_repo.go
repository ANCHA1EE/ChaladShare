package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"chaladshare_backend/internal/auth/models"
)

type AuthRepository interface {
	GetAllUsers() ([]models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	IsEmailTaken(email string) (bool, error)
	IsUsernameTaken(username string) (bool, error)
	CreateUser(email, username, passwordHash string) (*models.User, error)

	// ✅ เพิ่มของ reset password
	CreatePasswordReset(userID int, otpHash string, expiresAt time.Time) error
	GetLatestActivePasswordReset(userID int) (*models.PasswordReset, error)
	MarkPasswordResetUsed(resetID int) error
	MarkAllActivePasswordResetsUsed(userID int) error
	UpdateUserPasswordHash(userID int, passwordHash string) error
	// ✅ email verify (ก่อนสมัคร) 88
	CreateEmailVerification(email string, otpHash string, expiresAt time.Time) error
	GetLatestActiveEmailVerification(email string) (*models.EmailVerification, error)
	MarkEmailVerificationUsed(verifyID int) error
	MarkAllActiveEmailVerificationsUsed(email string) error

	// sessions refresh
	CreateSession(userID int, refreshHash string, expiresAt time.Time) (*models.AuthSession, error)
	GetSessionByRefresh(refreshHash string) (*models.AuthSession, error)
	RevokeSession(sessionID int) error
	RotateSession(oldSessionID int, newRefreshHash string, newExpiresAt time.Time) (*models.AuthSession, error)
	UpdateSessionLastUsed(sessionID int) error
}

type authRepository struct {
	db *sql.DB
}

// func สร้าง repository
func NewAuthRepository(db *sql.DB) AuthRepository {
	return &authRepository{db: db}
}

// GET ผู้ใช้ทั้งหมด เรียง id
func (r *authRepository) GetAllUsers() ([]models.User, error) {
	rows, err := r.db.Query(`
		SELECT user_id, email, username, user_created_at, user_status
		FROM users
		ORDER BY user_id
	`)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถดึงข้อมูลผู้ใช้ทั้งหมดได้: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.Username,
			&u.CreatedAt, &u.Status,
		); err != nil {
			return nil, fmt.Errorf("อ่านข้อมูลผู้ใช้ไม่สำเร็จ: %w", err)
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดระหว่างอ่านข้อมูล: %w", err)
	}
	return users, nil
}

// ผู้ใช้ตาม email
func (r *authRepository) GetUserByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(`
		SELECT user_id, email, username, password_hash, user_created_at, user_status
		FROM users
		WHERE LOWER(email) = LOWER($1)
	`, email).Scan(
		&u.ID, &u.Email, &u.Username, &u.PasswordHash,
		&u.CreatedAt, &u.Status,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("ไม่พบบัญชีผู้ใช้")
	} else if err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาด: %w", err)
	}
	return &u, nil
}

// สร้างผู้ใช้ใหม่
func (r *authRepository) CreateUser(email, username, passwordHash string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(`
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING user_id, email, username, user_created_at, user_status
	`, email, username, passwordHash).Scan(
		&u.ID, &u.Email, &u.Username,
		&u.CreatedAt, &u.Status,
	)

	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถสร้างผู้ใช้ใหม่ได้: %w", err)
	}

	return &u, nil
}
func (r *authRepository) CreatePasswordReset(userID int, otpHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO password_resets (reset_pass_user_id, otp_hash, reset_pass_expires_at, used_at)
		VALUES ($1, $2, $3, NULL)
	`, userID, otpHash, expiresAt)
	if err != nil {
		return fmt.Errorf("create password reset failed: %w", err)
	}
	return nil
}

func (r *authRepository) GetLatestActivePasswordReset(userID int) (*models.PasswordReset, error) {
	var pr models.PasswordReset
	err := r.db.QueryRow(`
		SELECT reset_pass_id, reset_pass_user_id, otp_hash, reset_pass_expires_at, used_at
		FROM password_resets
		WHERE reset_pass_user_id = $1
		  AND used_at IS NULL
		  AND reset_pass_expires_at > NOW()
		ORDER BY reset_pass_id DESC
		LIMIT 1
	`, userID).Scan(&pr.ID, &pr.UserID, &pr.OTPHash, &pr.ExpiresAt, &pr.UsedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("no active otp")
	}
	if err != nil {
		return nil, fmt.Errorf("get active otp failed: %w", err)
	}
	return &pr, nil
}

func (r *authRepository) MarkPasswordResetUsed(resetID int) error {
	_, err := r.db.Exec(`
		UPDATE password_resets
		SET used_at = NOW()
		WHERE reset_pass_id = $1
	`, resetID)
	if err != nil {
		return fmt.Errorf("mark reset used failed: %w", err)
	}
	return nil
}

// ปิด OTP เก่าทั้งหมดของ user (กันมีหลายอัน active)
func (r *authRepository) MarkAllActivePasswordResetsUsed(userID int) error {
	_, err := r.db.Exec(`
		UPDATE password_resets
		SET used_at = NOW()
		WHERE reset_pass_user_id = $1
		  AND used_at IS NULL
	`, userID)
	if err != nil {
		return fmt.Errorf("mark old resets used failed: %w", err)
	}
	return nil
}

func (r *authRepository) UpdateUserPasswordHash(userID int, passwordHash string) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET password_hash = $2
		WHERE user_id = $1
	`, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update password failed: %w", err)
	}
	return nil
}
func (r *authRepository) CreateEmailVerification(email string, otpHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO email_verifications (email, otp_hash, expires_at, used_at)
		VALUES ($1, $2, $3, NULL)
	`, email, otpHash, expiresAt)
	if err != nil {
		return fmt.Errorf("create email verification failed: %w", err)
	}
	return nil
}

func (r *authRepository) GetLatestActiveEmailVerification(email string) (*models.EmailVerification, error) {
	var ev models.EmailVerification
	err := r.db.QueryRow(`
		SELECT verify_id, email, otp_hash, expires_at, used_at, created_at
		FROM email_verifications
		WHERE lower(email) = lower($1)
		  AND used_at IS NULL
		  AND expires_at > NOW()
		ORDER BY verify_id DESC
		LIMIT 1
	`, email).Scan(&ev.ID, &ev.Email, &ev.OTPHash, &ev.ExpiresAt, &ev.UsedAt, &ev.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("no active otp")
	}
	if err != nil {
		return nil, fmt.Errorf("get active email otp failed: %w", err)
	}
	return &ev, nil
}

func (r *authRepository) MarkEmailVerificationUsed(verifyID int) error {
	_, err := r.db.Exec(`
		UPDATE email_verifications
		SET used_at = NOW()
		WHERE verify_id = $1
	`, verifyID)
	if err != nil {
		return fmt.Errorf("mark email verification used failed: %w", err)
	}
	return nil
}

// 88
func (r *authRepository) MarkAllActiveEmailVerificationsUsed(email string) error {
	_, err := r.db.Exec(`
		UPDATE email_verifications
		SET used_at = NOW()
		WHERE lower(email) = lower($1)
		  AND used_at IS NULL
	`, email)
	if err != nil {
		return fmt.Errorf("mark old email verifications used failed: %w", err)
	}
	return nil
}
func (r *authRepository) IsEmailTaken(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE LOWER(email) = LOWER($1)
		)
	`, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check email exists failed: %w", err)
	}
	return exists, nil
}

func (r *authRepository) IsUsernameTaken(username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE username_ci = LOWER($1)
		)
	`, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check username exists failed: %w", err)
	}
	return exists, nil
}

func (r *authRepository) CreateSession(userID int, refreshHash string, expiresAt time.Time) (*models.AuthSession, error) {
	var s models.AuthSession
	err := r.db.QueryRow(`
		INSERT INTO auth_sessions (
			session_user_id, refresh_token_hash, session_expires_at,
			created_at, last_used_at, revoked_at, replaced_by_session_id
		)
		VALUES ($1, $2, $3, NOW(), NOW(), NULL, NULL)
		RETURNING
			session_id, session_user_id, refresh_token_hash, session_expires_at,
			revoked_at, replaced_by_session_id, created_at, last_used_at
	`, userID, refreshHash, expiresAt).Scan(
		&s.SessionID, &s.UserID, &s.RefreshTokenHash, &s.ExpiresAt,
		&s.RevokedAt, &s.ReplacedByID, &s.CreatedAt, &s.LastUsedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create session failed: %w", err)
	}
	return &s, nil
}

func (r *authRepository) GetSessionByRefresh(refreshHash string) (*models.AuthSession, error) {
	var s models.AuthSession
	err := r.db.QueryRow(`
  SELECT
    session_id, session_user_id, refresh_token_hash, session_expires_at,
    revoked_at, replaced_by_session_id, created_at, last_used_at
  FROM auth_sessions
  WHERE refresh_token_hash = $1
    AND revoked_at IS NULL
    AND session_expires_at > NOW()
  ORDER BY session_id DESC
  LIMIT 1
`, refreshHash).Scan(
		&s.SessionID, &s.UserID, &s.RefreshTokenHash, &s.ExpiresAt,
		&s.RevokedAt, &s.ReplacedByID, &s.CreatedAt, &s.LastUsedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("session not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("get session by refresh failed: %w", err)
	}
	return &s, nil
}

func (r *authRepository) RevokeSession(sessionID int) error {
	_, err := r.db.Exec(`
		UPDATE auth_sessions
		SET revoked_at = NOW()
		WHERE session_id = $1
		  AND revoked_at IS NULL
	`, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session failed: %w", err)
	}
	return nil
}

func (r *authRepository) UpdateSessionLastUsed(sessionID int) error {
	_, err := r.db.Exec(`
		UPDATE auth_sessions
		SET last_used_at = NOW()
		WHERE session_id = $1
	`, sessionID)
	if err != nil {
		return fmt.Errorf("update session last used failed: %w", err)
	}
	return nil
}

func (r *authRepository) RotateSession(oldSessionID int, newRefreshHash string, newExpiresAt time.Time) (*models.AuthSession, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx failed: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// ล็อก session เก่า
	var userID int
	err = tx.QueryRow(`
		SELECT session_user_id
		FROM auth_sessions
		WHERE session_id = $1
		  AND revoked_at IS NULL
		  AND session_expires_at > NOW()
		FOR UPDATE
	`, oldSessionID).Scan(&userID)
	if err == sql.ErrNoRows {
		return nil, errors.New("old session invalid/expired")
	}
	if err != nil {
		return nil, fmt.Errorf("lock old session failed: %w", err)
	}

	// สร้าง session ใหม่
	var ns models.AuthSession
	err = tx.QueryRow(`
		INSERT INTO auth_sessions (
			session_user_id, refresh_token_hash, session_expires_at,
			created_at, last_used_at, revoked_at, replaced_by_session_id
		)
		VALUES ($1, $2, $3, NOW(), NOW(), NULL, NULL)
		RETURNING
			session_id, session_user_id, refresh_token_hash, session_expires_at,
			revoked_at, replaced_by_session_id, created_at, last_used_at
	`, userID, newRefreshHash, newExpiresAt).Scan(
		&ns.SessionID, &ns.UserID, &ns.RefreshTokenHash, &ns.ExpiresAt,
		&ns.RevokedAt, &ns.ReplacedByID, &ns.CreatedAt, &ns.LastUsedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert new session failed: %w", err)
	}

	// ปิด session เก่า แทนด้วย session ใหม่
	_, err = tx.Exec(`
		UPDATE auth_sessions
		SET revoked_at = NOW(),
		    replaced_by_session_id = $2
		WHERE session_id = $1
		  AND revoked_at IS NULL
	`, oldSessionID, ns.SessionID)
	if err != nil {
		return nil, fmt.Errorf("revoke old session failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx failed: %w", err)
	}
	return &ns, nil
}
