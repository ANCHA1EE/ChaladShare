package models

import (
	"encoding/json"
	"time"
)

const (
	FeatureQueued     = "queued"
	FeatureProcessing = "processing"
	FeatureDone       = "done"
	FeatureFailed     = "failed"
)

type DocumentFeature struct {
	DocumentID    int             `json:"document_id"`
	FeatureStatus string          `json:"feature_status"`
	StyleLabel    *string         `json:"style_label,omitempty"`
	StyleVector   json.RawMessage `json:"style_vector,omitempty"`
	ClusterID     *int            `json:"cluster_id,omitempty"`
	ErrorMessage  *string         `json:"error_message,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// ตอนสร้างแถวเริ่มต้น
type CreateQueuedInput struct {
	DocumentID int `json:"document_id"`
}

type SaveResult struct {
	DocumentID       int       `json:"document_id"`
	StyleLabel       string    `json:"style_label"`
	StyleVectorV16   []float64 `json:"style_vector_v16,omitempty"`
	ContentText      *string   `json:"content_text,omitempty"`
	ContentEmbedding []float64 `json:"content_embedding,omitempty"`
	ClusterID        *int      `json:"cluster_id,omitempty"`
}

func (df *DocumentFeature) VectorAsFloat64() ([]float64, error) {
	if len(df.StyleVector) == 0 {
		return nil, nil
	}
	var v []float64
	if err := json.Unmarshal(df.StyleVector, &v); err != nil {
		return nil, err
	}
	return v, nil
}
