package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"

	"chaladshare_backend/internal/posts/models"
)

type PostRepository interface {
	CreatePost(post *models.Post, tags []string) (int, error)
	UpdatePost(post *models.Post, tags []string) error
	DeletePost(postID int) error

	GetAllPosts() ([]models.PostResponse, error)
	GetPostByID(postID int) (*models.PostResponse, error)
	GetPostByIDForViewer(viewerID, postID int) (*models.PostResponse, error)
	GetFeedPosts(viewerID int) ([]models.PostResponse, error)
	GetPostOwnerID(postID int) (int, error)
	CountByUserID(userID int) (int, error)

	GetSavedPosts(userID int) ([]models.PostResponse, error)
	GetPopularPosts(viewerID, limit int) ([]models.PostResponse, error)
	SearchPosts(viewerID int, search string, page, size int) ([]models.PostResponse, int, error)
}

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) CreatePost(post *models.Post, tags []string) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if post.DocumentID == nil {
		return 0, fmt.Errorf("document_id is required")
	}
	var docArg interface{} = *post.DocumentID

	var coverArg interface{} = nil
	if post.CoverURL != nil {
		coverArg = *post.CoverURL
	}

	query := `INSERT INTO posts (post_author_user_id, post_title, post_description,
			  post_visibility, post_document_id, post_cover_url) 
			  SELECT $1, $2, $3, $4, $5, $6
			  FROM documents d
			  WHERE d.document_id = $5 AND d.document_user_id = $1
			  RETURNING post_id;`

	var postID int
	if err := tx.QueryRow(
		query,
		post.AuthorUserID, post.Title, post.Description,
		post.Visibility, docArg, coverArg,
	).Scan(&postID); err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("invalid document_id or not owned by user")
		}
		return 0, fmt.Errorf("create post: %w", err)
	}

	if len(tags) > 0 {
		upsertTag := `INSERT INTO tags (tag_name) VALUES ($1) ON CONFLICT (tag_name) DO UPDATE
					  SET tag_name = EXCLUDED.tag_name RETURNING tag_id;`

		link := `INSERT INTO post_tags (post_tag_post_id, post_tag_tag_id)
				 VALUES ($1, $2) ON CONFLICT DO NOTHING;`

		for _, t := range tags {
			var tagID int
			if err := tx.QueryRow(upsertTag, t).Scan(&tagID); err != nil {
				return 0, fmt.Errorf("upsert tag %q: %w", t, err)
			}
			if _, err := tx.Exec(link, postID, tagID); err != nil {
				return 0, fmt.Errorf("link tag %q: %w", t, err)
			}
		}
	}

	initStats := `INSERT INTO post_stats (post_stats_post_id, post_like_count, post_save_count)
				  VALUES ($1, 0, 0) ON CONFLICT DO NOTHING;`
	if _, err := tx.Exec(initStats, postID); err != nil {
		return 0, fmt.Errorf("init stats: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return postID, nil
}

func (r *postRepository) UpdatePost(post *models.Post, tags []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`UPDATE posts SET post_title = $1,
        				 post_description = $2, post_visibility = $3, post_updated_at = now()
    					 WHERE post_id = $4;`,
		post.Title, post.Description, post.Visibility, post.PostID)
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}

	if tags != nil {
		if _, err := tx.Exec(`DELETE FROM post_tags WHERE post_tag_post_id = $1`, post.PostID); err != nil {
			return fmt.Errorf("clear old tags: %w", err)
		}

		if len(tags) > 0 {
			upsertTag := `INSERT INTO tags (tag_name)
						  VALUES ($1) ON CONFLICT (tag_name) DO UPDATE
						  SET tag_name = EXCLUDED.tag_name
						  RETURNING tag_id;`

			link := `INSERT INTO post_tags (post_tag_post_id, post_tag_tag_id)
					 VALUES ($1, $2) ON CONFLICT DO NOTHING;`

			for _, t := range tags {
				var tagID int
				if err := tx.QueryRow(upsertTag, t).Scan(&tagID); err != nil {
					return fmt.Errorf("upsert tag %q: %w", t, err)
				}
				if _, err := tx.Exec(link, post.PostID, tagID); err != nil {
					return fmt.Errorf("link tag %q: %w", t, err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *postRepository) DeletePost(postID int) error {
	query := `DELETE FROM posts WHERE post_id = $1`
	res, err := r.db.Exec(query, postID)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *postRepository) GetAllPosts() ([]models.PostResponse, error) {
	query := `SELECT p.post_id, p.post_author_user_id, u.username AS author_name,
		p.post_title, p.post_description, p.post_visibility,
		p.post_document_id, p.post_created_at, p.post_updated_at,
		COALESCE(ps.post_like_count, 0) AS post_like_count,
		COALESCE(ps.post_save_count, 0) AS post_save_count,
		d.document_url AS document_file_url,
		d.document_name AS document_name,
		p.post_cover_url, up.avatar_url,
		ARRAY_REMOVE(ARRAY_AGG(DISTINCT t.tag_name), NULL) AS tags
	FROM posts p
	JOIN users u ON u.user_id = p.post_author_user_id
	LEFT JOIN post_stats ps ON ps.post_stats_post_id = p.post_id
	LEFT JOIN post_tags pt ON pt.post_tag_post_id = p.post_id
	LEFT JOIN tags t ON t.tag_id = pt.post_tag_tag_id
	LEFT JOIN documents d ON d.document_id = p.post_document_id
	LEFT JOIN user_profiles up ON up.profile_user_id = u.user_id
	GROUP BY p.post_id, u.username, ps.post_like_count, ps.post_save_count, d.document_url, d.document_name, p.post_cover_url, up.avatar_url
	ORDER BY p.post_created_at DESC;`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostResponse
	for rows.Next() {
		var (
			p         models.PostResponse
			tags      pq.StringArray
			fileURL   sql.NullString
			docName   sql.NullString
			coverURL  sql.NullString
			avatarURL sql.NullString
			docID     sql.NullInt64
		)

		if err := rows.Scan(
			&p.PostID, &p.AuthorID, &p.AuthorName,
			&p.Title, &p.Description, &p.Visibility,
			&docID, &p.CreatedAt, &p.UpdatedAt,
			&p.LikeCount, &p.SaveCount,
			&fileURL, &docName, &coverURL, &avatarURL, &tags,
		); err != nil {
			return nil, err
		}

		if docID.Valid {
			v := int(docID.Int64)
			p.DocumentID = &v
		}
		if fileURL.Valid {
			p.FileURL = &fileURL.String
		}
		if docName.Valid {
			p.DocumentName = &docName.String
		}
		if coverURL.Valid {
			p.CoverURL = &coverURL.String
		}
		if avatarURL.Valid {
			p.AvatarURL = &avatarURL.String
		}

		p.Tags = []string(tags)
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) GetFeedPosts(viewerID int) ([]models.PostResponse, error) {
	query := `
		SELECT p.post_id, p.post_author_user_id, u.username AS author_name,
			p.post_title, p.post_description, p.post_visibility,
			p.post_document_id, p.post_created_at, p.post_updated_at,
			COALESCE(ps.post_like_count, 0) AS post_like_count,
			COALESCE(ps.post_save_count, 0) AS post_save_count,
			d.document_url AS document_file_url,
			d.document_name AS document_name,
			p.post_cover_url, up.avatar_url,
			ARRAY_REMOVE(ARRAY_AGG(DISTINCT t.tag_name), NULL) AS tags,

			EXISTS (
				SELECT 1 FROM likes l
				WHERE l.like_user_id = $1 AND l.like_post_id = p.post_id
			) AS is_liked,
			EXISTS (
				SELECT 1 FROM saved_posts sp
				WHERE sp.save_user_id = $1 AND sp.save_post_id = p.post_id
			) AS is_saved

		FROM posts p
		JOIN users u ON u.user_id = p.post_author_user_id
		LEFT JOIN post_stats ps ON ps.post_stats_post_id = p.post_id
		LEFT JOIN post_tags pt ON pt.post_tag_post_id = p.post_id
		LEFT JOIN tags t ON t.tag_id = pt.post_tag_tag_id
		LEFT JOIN documents d ON d.document_id = p.post_document_id
		LEFT JOIN user_profiles up ON up.profile_user_id = u.user_id
		WHERE
			p.post_author_user_id = $1
			OR p.post_visibility = 'public'
			OR ( p.post_visibility = 'friends'
				AND EXISTS (
					SELECT 1
					FROM friendships f
					WHERE
						f.user_id  = LEAST(p.post_author_user_id, $1)
						AND f.friend_id = GREATEST(p.post_author_user_id, $1)
				)
			)
		GROUP BY p.post_id, u.username, ps.post_like_count, ps.post_save_count,
				 d.document_url, d.document_name, p.post_cover_url, up.avatar_url
		ORDER BY p.post_created_at DESC;
	`

	rows, err := r.db.Query(query, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostResponse
	for rows.Next() {
		var (
			p         models.PostResponse
			tags      pq.StringArray
			fileURL   sql.NullString
			docName   sql.NullString
			coverURL  sql.NullString
			avatarURL sql.NullString
			docID     sql.NullInt64
			isLiked   bool
			isSaved   bool
		)
		if err := rows.Scan(
			&p.PostID, &p.AuthorID, &p.AuthorName,
			&p.Title, &p.Description, &p.Visibility,
			&docID, &p.CreatedAt, &p.UpdatedAt,
			&p.LikeCount, &p.SaveCount,
			&fileURL, &docName, &coverURL, &avatarURL, &tags,
			&isLiked, &isSaved,
		); err != nil {
			return nil, err
		}

		if docID.Valid {
			v := int(docID.Int64)
			p.DocumentID = &v
		}
		if fileURL.Valid {
			p.FileURL = &fileURL.String
		}
		if docName.Valid {
			p.DocumentName = &docName.String
		}
		if coverURL.Valid {
			p.CoverURL = &coverURL.String
		}
		if avatarURL.Valid {
			p.AvatarURL = &avatarURL.String
		}

		p.IsLiked = isLiked
		p.IsSaved = isSaved

		p.Tags = []string(tags)
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) GetPostByID(postID int) (*models.PostResponse, error) {
	query := `SELECT p.post_id, p.post_author_user_id, u.username AS author_name,
		p.post_title, p.post_description, p.post_visibility, p.post_document_id,
		p.post_created_at, p.post_updated_at,
		COALESCE(ps.post_like_count, 0)  AS post_like_count,
		COALESCE(ps.post_save_count, 0)  AS post_save_count,
		d.document_url AS document_file_url,
		d.document_name AS document_name,
		p.post_cover_url, up.avatar_url,
		ARRAY_REMOVE(ARRAY_AGG(DISTINCT t.tag_name), NULL) AS tags
	FROM posts p
	JOIN users u ON u.user_id = p.post_author_user_id
	LEFT JOIN post_stats ps ON ps.post_stats_post_id = p.post_id
	LEFT JOIN post_tags pt ON pt.post_tag_post_id = p.post_id
	LEFT JOIN tags t ON t.tag_id = pt.post_tag_tag_id
	LEFT JOIN documents d ON d.document_id = p.post_document_id
	LEFT JOIN user_profiles up ON up.profile_user_id = u.user_id
	WHERE p.post_id = $1
	GROUP BY p.post_id, u.username, ps.post_like_count, ps.post_save_count, d.document_url, d.document_name, up.avatar_url;`

	row := r.db.QueryRow(query, postID)
	var (
		p         models.PostResponse
		tags      pq.StringArray
		fileURL   sql.NullString
		docName   sql.NullString
		coverURL  sql.NullString
		avatarURL sql.NullString
		docID     sql.NullInt64
	)

	if err := row.Scan(
		&p.PostID, &p.AuthorID, &p.AuthorName,
		&p.Title, &p.Description, &p.Visibility,
		&docID, &p.CreatedAt, &p.UpdatedAt,
		&p.LikeCount, &p.SaveCount,
		&fileURL, &docName, &coverURL, &avatarURL, &tags,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	if docID.Valid {
		v := int(docID.Int64)
		p.DocumentID = &v
	}
	if fileURL.Valid {
		p.FileURL = &fileURL.String
	}
	if docName.Valid {
		p.DocumentName = &docName.String
	}
	if coverURL.Valid {
		p.CoverURL = &coverURL.String
	}
	if avatarURL.Valid {
		p.AvatarURL = &avatarURL.String
	}

	p.Tags = []string(tags)
	return &p, nil
}

func (r *postRepository) GetPostByIDForViewer(viewerID, postID int) (*models.PostResponse, error) {
	query := `
	SELECT p.post_id, p.post_author_user_id, u.username AS author_name,
		p.post_title, p.post_description, p.post_visibility, p.post_document_id,
		p.post_created_at, p.post_updated_at,
		COALESCE(ps.post_like_count, 0) AS post_like_count,
		COALESCE(ps.post_save_count, 0) AS post_save_count,
		d.document_url AS document_file_url,
		d.document_name AS document_name,
		p.post_cover_url, up.avatar_url,
		ARRAY_REMOVE(ARRAY_AGG(DISTINCT t.tag_name), NULL) AS tags,

		EXISTS (
			SELECT 1 FROM likes l
			WHERE l.like_user_id = $1 AND l.like_post_id = p.post_id
		) AS is_liked,
		EXISTS (
			SELECT 1 FROM saved_posts sp
			WHERE sp.save_user_id = $1 AND sp.save_post_id = p.post_id
		) AS is_saved

	FROM posts p
	JOIN users u ON u.user_id = p.post_author_user_id
	LEFT JOIN post_stats ps ON ps.post_stats_post_id = p.post_id
	LEFT JOIN post_tags pt ON pt.post_tag_post_id = p.post_id
	LEFT JOIN tags t ON t.tag_id = pt.post_tag_tag_id
	LEFT JOIN documents d ON d.document_id = p.post_document_id
	LEFT JOIN user_profiles up ON up.profile_user_id = u.user_id
	WHERE p.post_id = $2
	GROUP BY p.post_id, u.username, ps.post_like_count, ps.post_save_count,
			 d.document_url, d.document_name, p.post_cover_url, up.avatar_url;
	`

	row := r.db.QueryRow(query, viewerID, postID)

	var (
		p         models.PostResponse
		tags      pq.StringArray
		fileURL   sql.NullString
		docName   sql.NullString
		coverURL  sql.NullString
		avatarURL sql.NullString
		docID     sql.NullInt64
		isLiked   bool
		isSaved   bool
	)

	if err := row.Scan(
		&p.PostID, &p.AuthorID, &p.AuthorName,
		&p.Title, &p.Description, &p.Visibility,
		&docID, &p.CreatedAt, &p.UpdatedAt,
		&p.LikeCount, &p.SaveCount,
		&fileURL, &docName, &coverURL, &avatarURL, &tags,
		&isLiked, &isSaved,
	); err != nil {
		return nil, err
	}

	if docID.Valid {
		v := int(docID.Int64)
		p.DocumentID = &v
	}
	if fileURL.Valid {
		p.FileURL = &fileURL.String
	}
	if docName.Valid {
		p.DocumentName = &docName.String
	}
	if coverURL.Valid {
		p.CoverURL = &coverURL.String
	}
	if avatarURL.Valid {
		p.AvatarURL = &avatarURL.String
	}

	p.Tags = []string(tags)
	p.IsLiked = isLiked
	p.IsSaved = isSaved
	return &p, nil
}

func (r *postRepository) GetPostOwnerID(postID int) (int, error) {
	const query = `SELECT post_author_user_id FROM posts WHERE post_id = $1`
	var owner int
	if err := r.db.QueryRow(query, postID).Scan(&owner); err != nil {
		return 0, err
	}
	return owner, nil
}

func (r *postRepository) CountByUserID(userID int) (int, error) {
	var cnt int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM posts WHERE post_author_user_id = $1`, userID).Scan(&cnt)
	return cnt, err
}

func (r *postRepository) GetSavedPosts(userID int) ([]models.PostResponse, error) {
	query := `
        SELECT p.post_id, p.post_author_user_id, u.username AS author_name,
               p.post_title, p.post_description, p.post_visibility,
               p.post_document_id, p.post_created_at, p.post_updated_at,
               COALESCE(ps.post_like_count, 0) AS post_like_count,
               COALESCE(ps.post_save_count, 0) AS post_save_count,
               d.document_url AS document_file_url,
			   d.document_name AS document_name,
               p.post_cover_url, up.avatar_url,
               ARRAY_REMOVE(ARRAY_AGG(DISTINCT t.tag_name), NULL) AS tags,

			   EXISTS (
					SELECT 1 FROM likes l
					WHERE l.like_user_id = $1 AND l.like_post_id = p.post_id
			   ) AS is_liked,
			   EXISTS (
					SELECT 1 FROM saved_posts sp2
					WHERE sp2.save_user_id = $1 AND sp2.save_post_id = p.post_id
			   ) AS is_saved

        FROM saved_posts sp
        JOIN posts p ON p.post_id = sp.save_post_id
        JOIN users u ON u.user_id = p.post_author_user_id
        LEFT JOIN post_stats ps ON ps.post_stats_post_id = p.post_id
        LEFT JOIN post_tags pt ON pt.post_tag_post_id = p.post_id
        LEFT JOIN tags t ON t.tag_id = pt.post_tag_tag_id
        LEFT JOIN documents d ON d.document_id = p.post_document_id
        LEFT JOIN user_profiles up ON up.profile_user_id = u.user_id
        WHERE sp.save_user_id = $1
          AND (
              p.post_author_user_id = $1
              OR p.post_visibility = 'public'
              OR ( p.post_visibility = 'friends'
                AND EXISTS (
                    SELECT 1
                    FROM friendships f
                    WHERE
                        f.user_id  = LEAST(p.post_author_user_id, $1)
                        AND f.friend_id = GREATEST(p.post_author_user_id, $1)
                )
              )
          )
        GROUP BY p.post_id, u.username, ps.post_like_count, ps.post_save_count,
                 d.document_url, d.document_name, p.post_cover_url, up.avatar_url
        ORDER BY p.post_created_at DESC;
    `

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostResponse
	for rows.Next() {
		var (
			p         models.PostResponse
			tags      pq.StringArray
			fileURL   sql.NullString
			docName   sql.NullString
			coverURL  sql.NullString
			avatarURL sql.NullString
			docID     sql.NullInt64
			isLiked   bool
			isSaved   bool
		)

		if err := rows.Scan(
			&p.PostID, &p.AuthorID, &p.AuthorName,
			&p.Title, &p.Description, &p.Visibility,
			&docID, &p.CreatedAt, &p.UpdatedAt,
			&p.LikeCount, &p.SaveCount,
			&fileURL, &docName, &coverURL, &avatarURL, &tags,
			&isLiked, &isSaved,
		); err != nil {
			return nil, err
		}
		if docID.Valid {
			v := int(docID.Int64)
			p.DocumentID = &v
		}
		if fileURL.Valid {
			p.FileURL = &fileURL.String
		}
		if docName.Valid {
			p.DocumentName = &docName.String
		}
		if coverURL.Valid {
			p.CoverURL = &coverURL.String
		}
		if avatarURL.Valid {
			p.AvatarURL = &avatarURL.String
		}
		p.IsLiked = isLiked
		p.IsSaved = isSaved
		p.Tags = []string(tags)
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) GetPopularPosts(viewerID, limit int) ([]models.PostResponse, error) {
	query := `
		SELECT p.post_id, p.post_author_user_id, u.username AS author_name,
			p.post_title, p.post_description, p.post_visibility,
			p.post_document_id, p.post_created_at, p.post_updated_at,
			COALESCE(ps.post_like_count, 0) AS post_like_count,
			COALESCE(ps.post_save_count, 0) AS post_save_count,
			d.document_url AS document_file_url,
			d.document_name AS document_name,
			p.post_cover_url, up.avatar_url,
			ARRAY_REMOVE(ARRAY_AGG(DISTINCT t.tag_name), NULL) AS tags,

			-- สถานะของผู้ชม (คนที่ล็อกอิน)
			EXISTS (
				SELECT 1 FROM likes l
				WHERE l.like_user_id = $1 AND l.like_post_id = p.post_id
			) AS is_liked,
			EXISTS (
				SELECT 1 FROM saved_posts sp
				WHERE sp.save_user_id = $1 AND sp.save_post_id = p.post_id
			) AS is_saved

		FROM posts p
		JOIN users u ON u.user_id = p.post_author_user_id
		LEFT JOIN post_stats ps ON ps.post_stats_post_id = p.post_id
		LEFT JOIN post_tags pt ON pt.post_tag_post_id = p.post_id
		LEFT JOIN tags t ON t.tag_id = pt.post_tag_tag_id
		LEFT JOIN documents d ON d.document_id = p.post_document_id
		LEFT JOIN user_profiles up ON up.profile_user_id = u.user_id
		WHERE p.post_visibility = 'public'
		GROUP BY p.post_id, u.username, ps.post_like_count, ps.post_save_count,
				 d.document_url, d.document_name, p.post_cover_url, up.avatar_url

		ORDER BY COALESCE(ps.post_like_count, 0) DESC, p.post_created_at DESC
		LIMIT $2;
	`

	rows, err := r.db.Query(query, viewerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.PostResponse
	for rows.Next() {
		var (
			p         models.PostResponse
			tags      pq.StringArray
			fileURL   sql.NullString
			docName   sql.NullString
			coverURL  sql.NullString
			avatarURL sql.NullString
			docID     sql.NullInt64
			isLiked   bool
			isSaved   bool
		)

		if err := rows.Scan(
			&p.PostID, &p.AuthorID, &p.AuthorName,
			&p.Title, &p.Description, &p.Visibility,
			&docID, &p.CreatedAt, &p.UpdatedAt,
			&p.LikeCount, &p.SaveCount,
			&fileURL, &docName, &coverURL, &avatarURL, &tags,
			&isLiked, &isSaved,
		); err != nil {
			return nil, err
		}

		if docID.Valid {
			v := int(docID.Int64)
			p.DocumentID = &v
		}
		if fileURL.Valid {
			p.FileURL = &fileURL.String
		}
		if docName.Valid {
			p.DocumentName = &docName.String
		}
		if coverURL.Valid {
			p.CoverURL = &coverURL.String
		}
		if avatarURL.Valid {
			p.AvatarURL = &avatarURL.String
		}

		p.Tags = []string(tags)
		p.IsLiked = isLiked
		p.IsSaved = isSaved

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepository) SearchPosts(viewerID int, search string, page, size int) ([]models.PostResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	search = strings.TrimSpace(search)
	pattern := "%" + search + "%"
	offset := (page - 1) * size

	// count post for page
	countQ := `
		SELECT COUNT(DISTINCT p.post_id)
		FROM posts p
		WHERE
			(
				p.post_author_user_id = $1
				OR p.post_visibility = 'public'
				OR (
					p.post_visibility = 'friends'
					AND EXISTS (
						SELECT 1
						FROM friendships f
						WHERE
							f.user_id = LEAST(p.post_author_user_id, $1)
							AND f.friend_id = GREATEST(p.post_author_user_id, $1)
					)
				)
			)
			AND (
				$2 = ''
				OR p.post_title ILIKE $3
				OR EXISTS (
					SELECT 1
					FROM post_tags pt2
					JOIN tags t2 ON t2.tag_id = pt2.post_tag_tag_id
					WHERE pt2.post_tag_post_id = p.post_id
					  AND t2.tag_name ILIKE $3
				)
			);
	`

	var total int
	if err := r.db.QueryRow(countQ, viewerID, search, pattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count search feed: %w", err)
	}

	// list query post
	listQ := `
		SELECT p.post_id, p.post_author_user_id, u.username AS author_name,
			p.post_title, p.post_description, p.post_visibility,
			p.post_document_id, p.post_created_at, p.post_updated_at,
			COALESCE(ps.post_like_count, 0) AS post_like_count,
			COALESCE(ps.post_save_count, 0) AS post_save_count,
			d.document_url AS document_file_url,
			d.document_name AS document_name,
			p.post_cover_url, up.avatar_url,
			ARRAY_REMOVE(ARRAY_AGG(DISTINCT t.tag_name), NULL) AS tags,

			EXISTS (
				SELECT 1 FROM likes l
				WHERE l.like_user_id = $1 AND l.like_post_id = p.post_id
			) AS is_liked,
			EXISTS (
				SELECT 1 FROM saved_posts sp
				WHERE sp.save_user_id = $1 AND sp.save_post_id = p.post_id
			) AS is_saved

		FROM posts p
		JOIN users u ON u.user_id = p.post_author_user_id
		LEFT JOIN post_stats ps ON ps.post_stats_post_id = p.post_id
		LEFT JOIN post_tags pt ON pt.post_tag_post_id = p.post_id
		LEFT JOIN tags t ON t.tag_id = pt.post_tag_tag_id
		LEFT JOIN documents d ON d.document_id = p.post_document_id
		LEFT JOIN user_profiles up ON up.profile_user_id = u.user_id

		WHERE
			(
				p.post_author_user_id = $1
				OR p.post_visibility = 'public'
				OR (
					p.post_visibility = 'friends'
					AND EXISTS (
						SELECT 1
						FROM friendships f
						WHERE
							f.user_id = LEAST(p.post_author_user_id, $1)
							AND f.friend_id = GREATEST(p.post_author_user_id, $1)
					)
				)
			)
			AND (
				$2 = ''
				OR p.post_title ILIKE $3
				OR EXISTS (
					SELECT 1
					FROM post_tags pt2
					JOIN tags t2 ON t2.tag_id = pt2.post_tag_tag_id
					WHERE pt2.post_tag_post_id = p.post_id
					  AND t2.tag_name ILIKE $3
				)
			)

		GROUP BY p.post_id, u.username, ps.post_like_count, ps.post_save_count,
				 d.document_url, d.document_name, p.post_cover_url, up.avatar_url
		ORDER BY p.post_created_at DESC
		LIMIT $4 OFFSET $5;
	`

	rows, err := r.db.Query(listQ, viewerID, search, pattern, size, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("search feed: %w", err)
	}
	defer rows.Close()

	var posts []models.PostResponse
	for rows.Next() {
		var (
			p         models.PostResponse
			tags      pq.StringArray
			fileURL   sql.NullString
			docName   sql.NullString
			coverURL  sql.NullString
			avatarURL sql.NullString
			docID     sql.NullInt64
			isLiked   bool
			isSaved   bool
		)

		if err := rows.Scan(
			&p.PostID, &p.AuthorID, &p.AuthorName,
			&p.Title, &p.Description, &p.Visibility,
			&docID, &p.CreatedAt, &p.UpdatedAt,
			&p.LikeCount, &p.SaveCount,
			&fileURL, &docName, &coverURL, &avatarURL, &tags,
			&isLiked, &isSaved,
		); err != nil {
			return nil, 0, err
		}

		if docID.Valid {
			v := int(docID.Int64)
			p.DocumentID = &v
		}
		if fileURL.Valid {
			p.FileURL = &fileURL.String
		}
		if docName.Valid {
			p.DocumentName = &docName.String
		}
		if coverURL.Valid {
			p.CoverURL = &coverURL.String
		}
		if avatarURL.Valid {
			p.AvatarURL = &avatarURL.String
		}

		p.Tags = []string(tags)
		p.IsLiked = isLiked
		p.IsSaved = isSaved

		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}
