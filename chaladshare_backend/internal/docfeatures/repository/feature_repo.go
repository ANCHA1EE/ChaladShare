package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"chaladshare_backend/internal/docfeatures/models"

	"github.com/pgvector/pgvector-go"
)

type DocFeaturesRepo interface {
	CreateQueued(documentID int) error
	MarkProcessing(documentID int) error
	SaveResult(input models.SaveResult) error
	MarkFailed(documentID int, msg string) error
	GetByDocumentID(documentID int) (*models.DocumentFeature, error)

	//
	ListVectors(label string, onlyUnclustered bool) ([]models.VectorItem, error)
	BatchUpdateClusters(updates []models.ClusterUpdate) (int, error)
	CountUnclustered(label string) (int, error)
	CountClusterable(label string) (int, error)
	DeleteByDocumentID(documentID int) error
}

type FeatureRepo struct {
	db *sql.DB
}

func NewFeatureRepo(db *sql.DB) DocFeaturesRepo {
	return &FeatureRepo{db: db}
}

func (r *FeatureRepo) CreateQueued(documentID int) error {
	q := `
		INSERT INTO document_features (document_id, feature_status)
		VALUES ($1, $2)
		ON CONFLICT (document_id) DO NOTHING;
	`
	_, err := r.db.Exec(q, documentID, models.FeatureQueued)
	return err
}

func (r *FeatureRepo) MarkProcessing(documentID int) error {
	q := `
		UPDATE document_features
		SET feature_status = $2, error_message = NULL
		WHERE document_id = $1;
	`
	_, err := r.db.Exec(q, documentID, models.FeatureProcessing)
	return err
}

func f64ToF32(a []float64) []float32 {
	out := make([]float32, len(a))
	for i, v := range a {
		out[i] = float32(v)
	}
	return out
}

func (r *FeatureRepo) SaveResult(input models.SaveResult) error {
	if len(input.StyleVectorV16) == 0 {
		return fmt.Errorf("empty style vector (len=0)")
	}

	vecJSON, err := json.Marshal(input.StyleVectorV16)
	if err != nil {
		return fmt.Errorf("marshal style vector: %w", err)
	}

	sv16 := pgvector.NewVector(f64ToF32(input.StyleVectorV16))

	var emb any = nil
	if len(input.ContentEmbedding) > 0 {
		emb = pgvector.NewVector(f64ToF32(input.ContentEmbedding))
	}

	q := `
		UPDATE document_features
		SET feature_status    = $2,
		    style_label       = $3,
		    style_vector_v16  = $4,
		    style_vector_raw  = $5::jsonb,
		    content_text      = $6,
		    content_embedding = $7,
		    cluster_id        = COALESCE($8, cluster_id),
		    error_message     = NULL
		WHERE document_id = $1;
	`
	_, err = r.db.Exec(q,
		input.DocumentID,
		models.FeatureDone,
		input.StyleLabel,
		sv16,
		vecJSON,
		input.ContentText,
		emb,
		input.ClusterID,
	)
	return err
}

func (r *FeatureRepo) MarkFailed(documentID int, msg string) error {
	q := `
		UPDATE document_features
		SET feature_status = $2, error_message = $3
		WHERE document_id = $1;
	`
	_, err := r.db.Exec(q, documentID, models.FeatureFailed, msg)
	return err
}

func (r *FeatureRepo) GetByDocumentID(documentID int) (*models.DocumentFeature, error) {
	q := `
		SELECT document_id, feature_status, style_label, style_vector_raw, cluster_id,
		       error_message, created_at, updated_at
		FROM document_features
		WHERE document_id = $1;
	`

	var out models.DocumentFeature
	err := r.db.QueryRow(q, documentID).Scan(
		&out.DocumentID,
		&out.FeatureStatus,
		&out.StyleLabel,
		&out.StyleVector,
		&out.ClusterID,
		&out.ErrorMessage,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// helper float32 > 64
func f32ToF64(a []float32) []float64 {
	out := make([]float64, len(a))
	for i, v := range a {
		out[i] = float64(v)
	}
	return out
}

func (r *FeatureRepo) ListVectors(label string, onlyUnclustered bool) ([]models.VectorItem, error) {
	q := `
		SELECT document_id, style_label, style_vector_v16
		FROM document_features
		WHERE feature_status = $1
		  AND style_label = $2
		  AND ($3 = false OR cluster_id IS NULL)
		ORDER BY document_id ASC;
	`

	rows, err := r.db.Query(q, models.FeatureDone, label, onlyUnclustered)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.VectorItem, 0, 128)

	for rows.Next() {
		var docID int
		var styleLabel string
		var v pgvector.Vector

		if err := rows.Scan(&docID, &styleLabel, &v); err != nil {
			return nil, err
		}

		item := models.VectorItem{
			DocumentID:     docID,
			StyleLabel:     styleLabel,
			StyleVectorV16: f32ToF64(v.Slice()),
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *FeatureRepo) BatchUpdateClusters(updates []models.ClusterUpdate) (int, error) {
	if len(updates) == 0 {
		return 0, nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.Prepare(`
	UPDATE document_features
	SET cluster_id = $2, cluster_updated_at = NOW()
	WHERE document_id = $1;
`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	updated := 0
	for _, u := range updates {
		if u.DocumentID <= 0 {
			continue
		}
		if _, err := stmt.Exec(u.DocumentID, u.ClusterID); err != nil {
			return 0, err
		}
		updated++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return updated, nil
}

func (r *FeatureRepo) CountUnclustered(label string) (int, error) {
	q := `
        SELECT COUNT(*)
        FROM document_features
        WHERE feature_status = $1
          AND style_label = $2
          AND style_vector_v16 IS NOT NULL
          AND cluster_id IS NULL;
    `
	var n int
	if err := r.db.QueryRow(q, models.FeatureDone, label).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *FeatureRepo) CountClusterable(label string) (int, error) {
	q := `
        SELECT COUNT(*)
        FROM document_features
        WHERE feature_status = $1
          AND style_label = $2
          AND style_vector_v16 IS NOT NULL;
    `
	var n int
	if err := r.db.QueryRow(q, models.FeatureDone, label).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (r *FeatureRepo) DeleteByDocumentID(documentID int) error {
	_, err := r.db.Exec(`DELETE FROM document_features WHERE document_id = $1`, documentID)
	return err
}
