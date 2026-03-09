package repository

import (
	"context"
	"database/sql"
	"errors"

	"chaladshare_backend/internal/friends/models"
)

type FriendRepository interface {
	// Follow
	InsertFollow(ctx context.Context, followerID, followedID int) error
	DeleteFollow(ctx context.Context, followerID, followedID int) error
	Following(ctx context.Context, followerID, followedID int) (bool, error)

	// Lists
	ListFriends(ctx context.Context, viewerID, userID int, search string, limit, offset int) ([]models.FriendItem, error)
	ListFollowers(ctx context.Context, viewerID, userID int, search string, limit, offset int) ([]models.FollowUser, error)
	ListFollowing(ctx context.Context, viewerID, userID int, search string, limit, offset int) ([]models.FollowUser, error)

	// Count
	CountFriends(ctx context.Context, userID int) (int, error)
	CountFollowers(ctx context.Context, userID int) (int, error)
	CountFollowing(ctx context.Context, userID int) (int, error)

	// Friend Requests
	CreateFriendRequest(ctx context.Context, requesterID, addresseeID int) (int, error)
	GetFriendRequest(ctx context.Context, requestID int) (*models.FriendRequest, error)
	GetPendingBetween(ctx context.Context, aID, bID int) (bool, error)
	ListIncomingRequests(ctx context.Context, addresseeID int, limit, offset int) ([]models.IncomingReqItem, error)
	ListOutgoingRequests(ctx context.Context, requesterID int, limit, offset int) ([]models.OutgoingReqItem, error)
	AcceptFriendRequest(ctx context.Context, requestID int, addresseeID int) error
	DeclineFriendRequest(ctx context.Context, requestID int, addresseeID int) error
	CancelFriendRequest(ctx context.Context, requestID int, requesterID int) error
	Unfriend(ctx context.Context, aID, bID int) error

	// Helper
	AreFriends(ctx context.Context, aID, bID int) (bool, error)

	/* 20-02 by ploy */

	// Search to Add friend
	SearchAddFriend(ctx context.Context, actorID int, search string, limit, offset int) ([]models.UserSearchItem, error)
	CountAddFriend(ctx context.Context, actorID int, search string) (int, error)

	/* 20-02 by ploy */

}

type friendrepo struct {
	db *sql.DB
}

func NewFriendRepository(db *sql.DB) FriendRepository {
	return &friendrepo{db: db}
}

func (r *friendrepo) InsertFollow(ctx context.Context, followerID, followedID int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO follows (follower_user_id, followed_user_id)
		VALUES ($1,$2) ON CONFLICT DO NOTHING`, followerID, followedID)
	return err
}

func (r *friendrepo) DeleteFollow(ctx context.Context, followerID, followedID int) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM follows WHERE follower_user_id=$1
		AND followed_user_id=$2`, followerID, followedID)
	return err
}

func (r *friendrepo) Following(ctx context.Context, followerID, followedID int) (bool, error) {
	var x int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1 FROM follows WHERE follower_user_id=$1 AND followed_user_id=$2
	`, followerID, followedID).Scan(&x)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *friendrepo) ListFriends(ctx context.Context, viewerID, userID int, search string, limit, offset int) ([]models.FriendItem, error) {
	const q = `
	WITH my_friends AS (
	  SELECT CASE WHEN f.user_id = $2 THEN f.friend_id ELSE f.user_id END AS friend_id
	  FROM friendships f
	  WHERE f.user_id = $2 OR f.friend_id = $2
	)
	SELECT
	  u.user_id                              AS user_id,
	  u.username                             AS username,
	  COALESCE(p.avatar_url,'')              AS avatar,
	  EXISTS (SELECT 1 FROM friendships fs
	          WHERE fs.user_id=LEAST($1::int, u.user_id) AND fs.friend_id=GREATEST($1::int, u.user_id)) AS is_friend,
	  EXISTS (SELECT 1 FROM follows f2
	          WHERE f2.follower_user_id=$1 AND f2.followed_user_id=u.user_id) AS is_following
	FROM my_friends mf
	JOIN users u ON u.user_id = mf.friend_id
	LEFT JOIN user_profiles p ON p.profile_user_id = u.user_id
	WHERE ($3 = '' OR u.username ILIKE '%'||$3||'%')
	ORDER BY u.username ASC, u.user_id ASC
	LIMIT $4 OFFSET $5;
	`
	rows, err := r.db.QueryContext(ctx, q, viewerID, userID, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.FriendItem
	for rows.Next() {
		var it models.FriendItem
		if err := rows.Scan(&it.UserID, &it.Username, &it.Avatar, &it.IsFriend, &it.IsFollowing); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *friendrepo) ListFollowers(ctx context.Context, viewerID, userID int, search string, limit, offset int) ([]models.FollowUser, error) {
	const q = `
	SELECT
	  u.user_id                             AS user_id,
	  u.username                            AS username,
	  COALESCE(p.avatar_url,'')             AS avatar,
	  EXISTS (SELECT 1 FROM friendships fs
	          WHERE fs.user_id=LEAST($1::int, u.user_id) AND fs.friend_id=GREATEST($1::int, u.user_id)) AS is_friend,
	  EXISTS (SELECT 1 FROM follows f2
	          WHERE f2.follower_user_id=$1 AND f2.followed_user_id=u.user_id) AS is_following
	FROM follows f
	JOIN users u ON u.user_id = f.follower_user_id
	LEFT JOIN user_profiles p ON p.profile_user_id = u.user_id
	WHERE f.followed_user_id = $2
	  AND ($3 = '' OR u.username ILIKE '%'||$3||'%')
	ORDER BY u.username ASC, u.user_id ASC
	LIMIT $4 OFFSET $5;
	`
	rows, err := r.db.QueryContext(ctx, q, viewerID, userID, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.FollowUser
	for rows.Next() {
		var it models.FollowUser
		if err := rows.Scan(&it.UserID, &it.Username, &it.Avatar, &it.IsFriend, &it.IsFollowing); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *friendrepo) ListFollowing(ctx context.Context, viewerID, userID int, search string, limit, offset int) ([]models.FollowUser, error) {
	const q = `
	SELECT
	  u.user_id                             AS user_id,
	  u.username                            AS username,
	  COALESCE(p.avatar_url,'')             AS avatar,
	  EXISTS (SELECT 1 FROM friendships fs
	          WHERE fs.user_id=LEAST($1::int, u.user_id) AND fs.friend_id=GREATEST($1::int, u.user_id)) AS is_friend,
	  EXISTS (SELECT 1 FROM follows f2
	          WHERE f2.follower_user_id=$1 AND f2.followed_user_id=u.user_id) AS is_following
	FROM follows f
	JOIN users u ON u.user_id = f.followed_user_id
	LEFT JOIN user_profiles p ON p.profile_user_id = u.user_id
	WHERE f.follower_user_id = $2
	  AND ($3 = '' OR u.username ILIKE '%'||$3||'%')
	ORDER BY u.username ASC, u.user_id ASC
	LIMIT $4 OFFSET $5;
	`
	rows, err := r.db.QueryContext(ctx, q, viewerID, userID, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.FollowUser
	for rows.Next() {
		var it models.FollowUser
		if err := rows.Scan(&it.UserID, &it.Username, &it.Avatar, &it.IsFriend, &it.IsFollowing); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *friendrepo) CountFriends(ctx context.Context, userID int) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM friendships
		WHERE user_id=$1 OR friend_id=$1
	`, userID).Scan(&n)
	return n, err
}

func (r *friendrepo) CountFollowers(ctx context.Context, userID int) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM follows
		WHERE followed_user_id=$1
	`, userID).Scan(&n)
	return n, err
}

func (r *friendrepo) CountFollowing(ctx context.Context, userID int) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM follows
		WHERE follower_user_id=$1
	`, userID).Scan(&n)
	return n, err
}

func (r *friendrepo) CreateFriendRequest(ctx context.Context, requesterID, addresseeID int) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO friend_requests (requester_user_id, addressee_user_id, request_status)
		VALUES ($1,$2,'pending')
		RETURNING request_id
	`, requesterID, addresseeID).Scan(&id)
	return id, err
}

func (r *friendrepo) GetFriendRequest(ctx context.Context, requestID int) (*models.FriendRequest, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT request_id, requester_user_id, addressee_user_id,
		       request_status, request_created_at, decided_at
		FROM friend_requests
		WHERE request_id=$1
	`, requestID)

	var fr models.FriendRequest
	if err := row.Scan(&fr.RequestID, &fr.RequesterUserID, &fr.AddresseeUserID,
		&fr.RequestStatus, &fr.RequestCreatedAt, &fr.DecidedAt); err != nil {
		return nil, err
	}
	return &fr, nil
}

func (r *friendrepo) GetPendingBetween(ctx context.Context, aID, bID int) (bool, error) {
	var x int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM friend_requests
		WHERE (
          (requester_user_id = $1 AND addressee_user_id = $2) OR
          (requester_user_id = $2 AND addressee_user_id = $1)
        )
        AND request_status = 'pending'::friend_request_status
        LIMIT 1
	`, aID, bID).Scan(&x)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *friendrepo) ListIncomingRequests(ctx context.Context, addresseeID int, limit, offset int) ([]models.IncomingReqItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		 SELECT fr.request_id,
               fr.requester_user_id,
               fr.request_created_at AS requested_at,
               u.username,
               COALESCE(p.avatar_url,'') AS avatar
        FROM friend_requests fr
        JOIN users u ON u.user_id = fr.requester_user_id
        LEFT JOIN user_profiles p ON p.profile_user_id = u.user_id
        WHERE fr.addressee_user_id = $1
          AND fr.request_status = 'pending'::friend_request_status
        ORDER BY fr.request_created_at DESC
        LIMIT $2 OFFSET $3
	`, addresseeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.IncomingReqItem
	for rows.Next() {
		var it models.IncomingReqItem
		if err := rows.Scan(&it.RequestID, &it.RequesterUserID, &it.RequestedAt, &it.Username, &it.Avatar); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *friendrepo) ListOutgoingRequests(ctx context.Context, requesterID int, limit, offset int) ([]models.OutgoingReqItem, error) {
	rows, err := r.db.QueryContext(ctx, `
  SELECT fr.request_id,
               fr.addressee_user_id      AS target_user_id,
               fr.request_created_at     AS requested_at,
               u.username,
               COALESCE(p.avatar_url,'') AS avatar
        FROM friend_requests fr
        JOIN users u ON u.user_id = fr.addressee_user_id
        LEFT JOIN user_profiles p ON p.profile_user_id = u.user_id
        WHERE fr.requester_user_id = $1
          AND fr.request_status = 'pending'::friend_request_status
        ORDER BY fr.request_created_at DESC
        LIMIT $2 OFFSET $3
`, requesterID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.OutgoingReqItem
	for rows.Next() {
		var it models.OutgoingReqItem
		if err := rows.Scan(&it.RequestID, &it.TargetUserID, &it.RequestedAt, &it.Username, &it.Avatar); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *friendrepo) AcceptFriendRequest(ctx context.Context, requestID int, addresseeID int) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1) mark accepted (ต้องเป็นผู้รับ และยัง pending)
	res, err := tx.ExecContext(ctx, `
		UPDATE friend_requests
		SET request_status='accepted'::friend_request_status,
		 	decided_at=now()
		WHERE request_id=$1 
			AND addressee_user_id=$2 
			AND request_status='pending'::friend_request_status
	`, requestID, addresseeID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}

	var requesterID int
	if err = tx.QueryRowContext(ctx, `
		SELECT requester_user_id FROM friend_requests WHERE request_id=$1
	`, requestID).Scan(&requesterID); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO friendships (user_id, friend_id)
		VALUES (LEAST($1::int, $2::int), GREATEST($1::int, $2::int))
		ON CONFLICT DO NOTHING
	`, addresseeID, requesterID); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO follows (follower_user_id, followed_user_id)
		VALUES ($1,$2) ON CONFLICT DO NOTHING
	`, addresseeID, requesterID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO follows (follower_user_id, followed_user_id)
		VALUES ($1,$2) ON CONFLICT DO NOTHING
	`, requesterID, addresseeID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *friendrepo) DeclineFriendRequest(ctx context.Context, requestID int, addresseeID int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE friend_requests
		SET request_status='declined'::friend_request_status,
			decided_at=now()
		WHERE request_id=$1 
			AND addressee_user_id=$2 
			AND request_status='pending'::friend_request_status
	`, requestID, addresseeID)
	return err
}

func (r *friendrepo) CancelFriendRequest(ctx context.Context, requestID int, requesterID int) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM friend_requests
		WHERE request_id=$1 
		AND requester_user_id=$2 
		AND request_status='pending'::friend_request_status
	`, requestID, requesterID)
	return err
}

// Unfriend: TX = delete friendship + delete follows A↔B
func (r *friendrepo) Unfriend(ctx context.Context, aID, bID int) (err error) {
	if aID == bID {
		return errors.New("cannot unfriend yourself")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `
		DELETE FROM friendships
		WHERE user_id=LEAST($1::int, $2::int) AND friend_id=GREATEST($1::int, $2::int)
	`, aID, bID); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, `
		DELETE FROM follows
		WHERE (follower_user_id=$1 AND followed_user_id=$2)
		   OR (follower_user_id=$2 AND followed_user_id=$1)
	`, aID, bID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *friendrepo) AreFriends(ctx context.Context, aID, bID int) (bool, error) {
	var x int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1 FROM friendships
		WHERE user_id=LEAST($1::int, $2::int) AND friend_id=GREATEST($1::int, $2::int)
	`, aID, bID).Scan(&x)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (r *friendrepo) SearchAddFriend(ctx context.Context, actorID int, search string, limit, offset int) ([]models.UserSearchItem, error) {
	const q = `
	SELECT
	  u.user_id,
	  u.username,
	  COALESCE(p.avatar_url,'') AS avatar,

	  -- case 2: เป็นเพื่อนแล้ว
	  EXISTS (
		SELECT 1 FROM friendships fs
		WHERE fs.user_id = LEAST($1::int, u.user_id)
		  AND fs.friend_id = GREATEST($1::int, u.user_id)
	  ) AS is_friend,

	  -- case 3: pending (ถ้ามี)
	  COALESCE(fr.request_id, 0) AS request_id,
	  COALESCE(fr.request_status::text, '') AS request_status,
	  CASE
	    WHEN fr.request_status = 'pending'::friend_request_status
	         AND fr.addressee_user_id = $1 THEN 'incoming'
	    WHEN fr.request_status = 'pending'::friend_request_status
	         AND fr.requester_user_id = $1 THEN 'outgoing'
	    ELSE 'none'
	  END AS request_direction

	FROM users u
	LEFT JOIN user_profiles p ON p.profile_user_id = u.user_id

	-- ดึง pending ระหว่าง actor กับ user 
	LEFT JOIN LATERAL (
	  SELECT fr2.request_id, fr2.request_status, fr2.requester_user_id, fr2.addressee_user_id
	  FROM friend_requests fr2
	  WHERE (
	    (fr2.requester_user_id = $1 AND fr2.addressee_user_id = u.user_id)
	    OR
	    (fr2.requester_user_id = u.user_id AND fr2.addressee_user_id = $1)
	  )
	  AND fr2.request_status = 'pending'::friend_request_status
	  ORDER BY fr2.request_created_at DESC
	  LIMIT 1
	) fr ON true

	WHERE u.user_id <> $1
	  AND ($2 = '' OR u.username_ci LIKE '%' || lower($2) || '%')

	ORDER BY u.username ASC, u.user_id ASC
	LIMIT $3 OFFSET $4;
	`

	rows, err := r.db.QueryContext(ctx, q, actorID, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.UserSearchItem{}
	for rows.Next() {
		var it models.UserSearchItem
		if err := rows.Scan(
			&it.UserID,
			&it.Username,
			&it.Avatar,
			&it.IsFriend,
			&it.RequestID,
			&it.RequestStatus,
			&it.RequestDirection,
		); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (r *friendrepo) CountAddFriend(ctx context.Context, actorID int, search string) (int, error) {
	const q = `
	SELECT COUNT(*)
	FROM users u
	WHERE u.user_id <> $1
	  AND ($2 = '' OR u.username_ci LIKE '%' || lower($2) || '%');
	`

	var n int
	err := r.db.QueryRowContext(ctx, q, actorID, search).Scan(&n)
	return n, err
}
