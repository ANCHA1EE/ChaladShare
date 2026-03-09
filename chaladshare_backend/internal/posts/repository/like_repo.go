package repository

import (
	"database/sql"
	"fmt"
)

type LikeRepository interface {
	LikePost(userID, postID int) error
	UnlikePost(userID, postID int) error
	IsPostLiked(userID, postID int) (bool, error)
	UpdateLikeCount(postID int) error
	LikeCount(postID int) (int, error)
}

type likeRepository struct {
	db *sql.DB
}

func NewLikeRepository(db *sql.DB) LikeRepository {
	return &likeRepository{db: db}
}

// กด like
func (r *likeRepository) LikePost(userID, postID int) error {
	query := `
		INSERT INTO likes (like_user_id, like_post_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;
	`
	if _, err := r.db.Exec(query, userID, postID); err != nil {
		return fmt.Errorf("failed to like post: %v", err)
	}
	return r.UpdateLikeCount(postID)
}

// ยกเลิก like
func (r *likeRepository) UnlikePost(userID, postID int) error {
	query := `DELETE FROM likes WHERE like_user_id=$1 AND like_post_id=$2`
	if _, err := r.db.Exec(query, userID, postID); err != nil {
		return fmt.Errorf("failed to unlike post: %v", err)
	}
	return r.UpdateLikeCount(postID)
}

// โพสต์ถูกกดไลก์หรือยัง
func (r *likeRepository) IsPostLiked(userID, postID int) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1 FROM likes
            WHERE like_user_id=$1 AND like_post_id=$2
        )
    `
	var liked bool
	if err := r.db.QueryRow(query, userID, postID).Scan(&liked); err != nil {
		return false, err
	}
	return liked, nil
}

// อัปเดตจำนวนไลก์ใน post_stat
func (r *likeRepository) UpdateLikeCount(postID int) error {
	query := `
        INSERT INTO post_stats (post_stats_post_id, post_like_count, post_last_activity_at)
        VALUES (
            $1,
            (SELECT COUNT(*) FROM likes WHERE like_post_id = $1),
            NOW()
        )
        ON CONFLICT (post_stats_post_id)
        DO UPDATE SET
            post_like_count       = EXCLUDED.post_like_count,
            post_last_activity_at = EXCLUDED.post_last_activity_at;
    `
	_, err := r.db.Exec(query, postID)
	return err
}

func (r *likeRepository) LikeCount(postID int) (int, error) {
	query := `SELECT COUNT(*) FROM likes WHERE like_post_id = $1`
	var count int
	if err := r.db.QueryRow(query, postID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get like count: %v", err)
	}
	return count, nil
}
