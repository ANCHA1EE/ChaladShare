package repository

import (
	"context"
	"database/sql"
	"errors"

	"chaladshare_backend/internal/users/models"
)

type UserRepository interface {
	GetOwnProfile(ctx context.Context, userID int) (*models.OwnProfileResponse, error)
	GetViewedUserProfile(ctx context.Context, userID int) (*models.ViewedUserProfileResponse, error)
	UpdateOwnProfile(ctx context.Context, userID int, req *models.UpdateOwnProfileRequest) error

	GetPasswordHash(ctx context.Context, userID int) (string, error)
	UpdatePasswordHash(ctx context.Context, userID int, newHash string) error
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

// owner profile
func (r *userRepo) GetOwnProfile(ctx context.Context, userID int) (*models.OwnProfileResponse, error) {
	query := `
			SELECT
			u.user_id, u.email, u.username,
			u.user_status,
			to_char(u.user_created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS user_created_at,
			p.avatar_url,
			p.avatar_storage,
			p.bio
			FROM users u
			LEFT JOIN user_profiles p ON p.profile_user_id = u.user_id
			WHERE u.user_id = $1
			`
	row := r.db.QueryRowContext(ctx, query, userID)

	var j models.UserProfile
	if err := row.Scan(
		&j.UserID, &j.Email, &j.Username, &j.Status, &j.CreatedAt,
		&j.AvatarURL, &j.AvatarStore, &j.Bio,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	resp := j.ToOwnProfileResponse()
	return &resp, nil
}

// viewed profile
func (r *userRepo) GetViewedUserProfile(ctx context.Context, userID int) (*models.ViewedUserProfileResponse, error) {
	query := `
			SELECT
			u.user_id, u.username,
			p.avatar_url, p.avatar_storage, p.bio
			FROM users u
			LEFT JOIN user_profiles p ON p.profile_user_id = u.user_id
			WHERE u.user_id = $1
			`
	row := r.db.QueryRowContext(ctx, query, userID)

	var j models.UserProfile
	if err := row.Scan(
		&j.UserID, &j.Username,
		&j.AvatarURL, &j.AvatarStore, &j.Bio); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	resp := j.ToViewedUserProfileResponse()
	return &resp, nil
}

func (r *userRepo) UpdateOwnProfile(ctx context.Context, userID int, req *models.UpdateOwnProfileRequest) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	if req.Username != nil {
		if _, err = tx.ExecContext(ctx,
			`UPDATE users SET username = $1 WHERE user_id = $2`,
			*req.Username, userID,
		); err != nil {
			return err // ถ้าชื่อซ้ำ/ติด unique constraint จะเด้งจากตรงนี้
		}
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO user_profiles(profile_user_id)
		VALUES ($1)
		ON CONFLICT (profile_user_id) DO NOTHING
	`, userID); err != nil {
		return err
	}

	if req.AvatarURL != nil || req.AvatarStore != nil || req.Bio != nil {
		if _, err = tx.ExecContext(ctx, `
			UPDATE user_profiles
			SET
				avatar_url     = COALESCE($1, avatar_url),
				avatar_storage = COALESCE($2, avatar_storage),
				bio            = COALESCE($3, bio),
				updated_at     = now()
			WHERE profile_user_id = $4
		`, req.AvatarURL, req.AvatarStore, req.Bio, userID); err != nil {
			return err
		}
	}

	return nil
}
func (r *userRepo) GetPasswordHash(ctx context.Context, userID int) (string, error) {
	var hash string
	err := r.db.QueryRowContext(ctx,
		`SELECT password_hash FROM users WHERE user_id = $1`,
		userID,
	).Scan(&hash)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func (r *userRepo) UpdatePasswordHash(ctx context.Context, userID int, newHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET password_hash = $1 WHERE user_id = $2`,
		newHash, userID,
	)
	return err
}
