package repository

import (
	"database/sql"
	"fmt"
	"strings"

	postmodels "chaladshare_backend/internal/posts/models"
	recmodels "chaladshare_backend/internal/recommend/models"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

type SeedPair struct {
	Label string
	CID   int
}

type RecommendRepo interface {
	ListSeedsFromLikes(userID int, limit int) ([]recmodels.SeedItem, []SeedPair, error)
	ListCandidatesBySeedPairs(userID int, pairs []SeedPair, limit int) ([]recmodels.CandidateItem, error)
	ReplaceUserRecommendations(userID int, recs []recmodels.ColabRecommendItem) error

	CountUserLikes(userID int) (int, error)
	ClearUserRecommendations(userID int) error

	ListRecommendedPosts(viewerID int, limit int) ([]postmodels.PostResponse, error)
}

type repo struct {
	db *sql.DB
}

func NewRecommendRepo(db *sql.DB) RecommendRepo {
	return &repo{db: db}
}

func f32ToF64(a []float32) []float64 {
	out := make([]float64, len(a))
	for i, v := range a {
		out[i] = float64(v)
	}
	return out
}

func (r *repo) ListSeedsFromLikes(userID int, limit int) ([]recmodels.SeedItem, []SeedPair, error) {
	if limit <= 0 {
		limit = 5
	}

	q := `
WITH liked_docs AS (
  SELECT
    p.post_document_id AS document_id,
    MAX(l.like_created_at) AS liked_at
  FROM likes l
  JOIN posts p ON p.post_id = l.like_post_id
  WHERE l.like_user_id = $1
    AND p.post_document_id IS NOT NULL
  GROUP BY p.post_document_id
  ORDER BY MAX(l.like_created_at) DESC
  LIMIT $2
)
SELECT
  ld.document_id,
  df.style_label,
  df.cluster_id,
  df.style_vector_v16
FROM liked_docs ld
JOIN document_features df ON df.document_id = ld.document_id
WHERE df.feature_status = 'done'
  AND df.style_vector_v16 IS NOT NULL
  AND df.cluster_id IS NOT NULL
  AND df.cluster_id >= 0
  AND df.style_label IN ('typed','handwritten');
`

	rows, err := r.db.Query(q, userID, limit)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var seeds []recmodels.SeedItem
	uniqPairs := map[string]SeedPair{}

	for rows.Next() {
		var docID int
		var label string
		var cid int
		var v pgvector.Vector

		if err := rows.Scan(&docID, &label, &cid, &v); err != nil {
			return nil, nil, err
		}

		seeds = append(seeds, recmodels.SeedItem{
			DocumentID:     docID,
			StyleLabel:     label,
			ClusterID:      cid,
			StyleVectorV16: f32ToF64(v.Slice()),
		})

		key := fmt.Sprintf("%s:%d", label, cid)
		uniqPairs[key] = SeedPair{Label: label, CID: cid}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	pairs := make([]SeedPair, 0, len(uniqPairs))
	for _, p := range uniqPairs {
		pairs = append(pairs, p)
	}
	return seeds, pairs, nil
}

func (r *repo) ListCandidatesBySeedPairs(userID int, pairs []SeedPair, limit int) ([]recmodels.CandidateItem, error) {
	if limit <= 0 {
		limit = 800
	}
	if len(pairs) == 0 {
		return []recmodels.CandidateItem{}, nil
	}

	uniqLabels := map[string]struct{}{}
	for _, p := range pairs {
		if p.Label != "" {
			uniqLabels[p.Label] = struct{}{}
		}
	}
	if len(uniqLabels) == 0 {
		return []recmodels.CandidateItem{}, nil
	}

	args := []any{userID}
	var b strings.Builder

	b.WriteString(`
WITH seed_labels(style_label) AS (
  VALUES
`)

	argPos := 2
	i := 0
	for label := range uniqLabels {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(fmt.Sprintf(" ($%d::text)", argPos))
		args = append(args, label)
		argPos++
		i++
	}

	b.WriteString(`
)
SELECT
  p.post_id,
  p.post_document_id AS document_id,
  df.style_label,
  df.cluster_id,
  df.style_vector_v16
FROM posts p
JOIN document_features df ON df.document_id = p.post_document_id
WHERE p.post_document_id IS NOT NULL
  AND df.feature_status = 'done'
  AND df.style_vector_v16 IS NOT NULL
  AND df.cluster_id IS NOT NULL
  AND df.cluster_id >= 0
  AND df.style_label IN ('typed','handwritten')
  AND p.post_author_user_id <> $1
  AND NOT EXISTS (
    SELECT 1 FROM likes l2
    WHERE l2.like_user_id = $1
      AND l2.like_post_id = p.post_id
  )
  AND (
    p.post_visibility = 'public'
    OR (
      p.post_visibility = 'friends'
      AND EXISTS (
        SELECT 1 FROM friendships f
        WHERE f.user_id = LEAST($1, p.post_author_user_id)
          AND f.friend_id = GREATEST($1, p.post_author_user_id)
      )
    )
  )
  AND EXISTS (
    SELECT 1 FROM seed_labels sl
    WHERE sl.style_label = df.style_label
  )
ORDER BY p.post_created_at DESC
LIMIT ` + fmt.Sprintf("%d", limit) + `;
`)

	rows, err := r.db.Query(b.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]recmodels.CandidateItem, 0, 256)
	for rows.Next() {
		var postID, docID int
		var label string
		var cid int
		var v pgvector.Vector

		if err := rows.Scan(&postID, &docID, &label, &cid, &v); err != nil {
			return nil, err
		}

		out = append(out, recmodels.CandidateItem{
			PostID:         postID,
			DocumentID:     docID,
			StyleLabel:     label,
			ClusterID:      cid,
			StyleVectorV16: f32ToF64(v.Slice()),
		})
	}
	return out, rows.Err()
}

func (r *repo) ReplaceUserRecommendations(userID int, recs []recmodels.ColabRecommendItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM recommendations WHERE rec_user_id = $1;`, userID); err != nil {
		return err
	}

	if len(recs) == 0 {
		return tx.Commit()
	}

	stmt, err := tx.Prepare(`
INSERT INTO recommendations (rec_user_id, rec_post_id, score, seed_document_id, created_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (rec_user_id, rec_post_id)
DO UPDATE SET score = EXCLUDED.score, seed_document_id = EXCLUDED.seed_document_id, created_at = NOW();
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, it := range recs {
		var seed any = nil
		if it.MatchedSeedDocumentID != nil {
			seed = *it.MatchedSeedDocumentID
		}
		if _, err := stmt.Exec(userID, it.PostID, it.Score, seed); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *repo) CountUserLikes(userID int) (int, error) {
	var n int
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM likes
		WHERE like_user_id = $1;
	`, userID).Scan(&n)
	return n, err
}

func (r *repo) ClearUserRecommendations(userID int) error {
	_, err := r.db.Exec(`
		DELETE FROM recommendations
		WHERE rec_user_id = $1;
	`, userID)
	return err
}

func (r *repo) ListRecommendedPosts(viewerID int, limit int) ([]postmodels.PostResponse, error) {
	if limit <= 0 {
		limit = 3
	}

	q := `
WITH rec AS (
    SELECT rec_post_id, score, created_at
    FROM recommendations
    WHERE rec_user_id = $1
    ORDER BY score DESC, created_at DESC
    LIMIT $2
)
SELECT
    p.post_id, p.post_author_user_id, u.username AS author_name,
    p.post_title, p.post_description, p.post_visibility,
    p.post_document_id, p.post_created_at, p.post_updated_at,

    COALESCE(ps.post_like_count, 0) AS post_like_count,
    COALESCE(ps.post_save_count, 0) AS post_save_count,

    d.document_url  AS document_file_url,
    d.document_name AS document_name,

    p.post_cover_url,
    up.avatar_url,

    ARRAY_REMOVE(ARRAY_AGG(DISTINCT t.tag_name), NULL) AS tags,

    EXISTS (
        SELECT 1 FROM likes l
        WHERE l.like_user_id = $1 AND l.like_post_id = p.post_id
    ) AS is_liked,
    EXISTS (
        SELECT 1 FROM saved_posts sp
        WHERE sp.save_user_id = $1 AND sp.save_post_id = p.post_id
    ) AS is_saved

FROM rec
JOIN posts p ON p.post_id = rec.rec_post_id
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
                    f.user_id  = LEAST(p.post_author_user_id, $1)
                    AND f.friend_id = GREATEST(p.post_author_user_id, $1)
            )
        )
    )

GROUP BY
    p.post_id, u.username,
    ps.post_like_count, ps.post_save_count,
    d.document_url, d.document_name,
    p.post_cover_url, up.avatar_url,
    rec.score, rec.created_at

ORDER BY rec.score DESC, rec.created_at DESC;
`

	rows, err := r.db.Query(q, viewerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]postmodels.PostResponse, 0, limit)

	for rows.Next() {
		var (
			p         postmodels.PostResponse
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

		out = append(out, p)
	}

	return out, rows.Err()
}
