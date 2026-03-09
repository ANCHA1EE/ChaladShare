package repository

import (
	"database/sql"
	"fmt"
)

type SaveRepository interface {
	SavePost(userID, postID int) error
	UnsavePost(userID, postID int) error
	IsPostSaved(userID, postID int) (bool, error)
	UpdateSaveCount(postID int) error
	SaveCount(postID int) (int, error)
}

type saveRepository struct {
	db *sql.DB
}

func NewSaveRepository(db *sql.DB) SaveRepository {
	return &saveRepository{db: db}
}

// บันทึกโพสต์
func (r *saveRepository) SavePost(userID, postID int) error {
	query := `
		INSERT INTO saved_posts (save_user_id, save_post_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;
	`
	if _, err := r.db.Exec(query, userID, postID); err != nil {
		return fmt.Errorf("failed to save post: %v", err)
	}
	return r.UpdateSaveCount(postID)
}

// ยกเลิกบันทึกโพสต์
func (r *saveRepository) UnsavePost(userID, postID int) error {
	query := `DELETE FROM saved_posts WHERE save_user_id=$1 AND save_post_id=$2`
	if _, err := r.db.Exec(query, userID, postID); err != nil {
		return fmt.Errorf("failed to unsave post: %v", err)
	}
	return r.UpdateSaveCount(postID)
}

// ตรวจสอบว่าเคยถูกบันทึกหรือยัง
func (r *saveRepository) IsPostSaved(userID, postID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM saved_posts
			WHERE save_user_id=$1 AND save_post_id=$2
		)
	`
	var saved bool
	if err := r.db.QueryRow(query, userID, postID).Scan(&saved); err != nil {
		return false, err
	}
	return saved, nil
}

// อัปเดตจำนวนบันทึกใน post_stat
func (r *saveRepository) UpdateSaveCount(postID int) error {
	query := `
		INSERT INTO post_stats (post_stats_post_id, post_save_count, post_last_activity_at)
		VALUES (
			$1,
			(SELECT COUNT(*) FROM saved_posts WHERE save_post_id = $1),
			NOW()
		)
		ON CONFLICT (post_stats_post_id)
		DO UPDATE SET
			post_save_count       = EXCLUDED.post_save_count,
			post_last_activity_at = EXCLUDED.post_last_activity_at;
	`
	_, err := r.db.Exec(query, postID)
	return err
}

// จำนวน save โพสต์
func (r *saveRepository) SaveCount(postID int) (int, error) {
	query := `SELECT COUNT(*) FROM saved_posts WHERE save_post_id = $1`
	var count int
	if err := r.db.QueryRow(query, postID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get save count: %v", err)
	}
	return count, nil
}
