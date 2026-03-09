package connect

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"
)

type ExtractResp struct {
	DocumentID         int       `json:"document_id"`
	StyleLabel         *string   `json:"style_label"`
	StyleVectorV16     []float64 `json:"style_vector_v16"`
	StyleVectorV16Norm []float64 `json:"style_vector_v16_norm,omitempty"`
	StyleVector        []float64 `json:"style_vector,omitempty"`
	ContentText        string    `json:"content_text"`
	Embedding          []float64 `json:"content_embedding"`
	EmbeddingAlt       []float64 `json:"embedding,omitempty"`
	ClusterID          *int      `json:"cluster_id,omitempty"`
}

func (c *Client) ExtractFeatures(documentID int, pdfPath string) (*ExtractResp, error) {
	start := time.Now()

	//context timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.ExtractTimeout)
	defer cancel()

	//ส่งไฟล์ผ่าน helper ใน client.go
	resp, err := c.postPDFWithField(ctx, "/extract", documentID, pdfPath, "file")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//เช็ค status code
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("extract status %d: %s", resp.StatusCode, string(b))
	}

	//decode JSON
	var out ExtractResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	if len(out.Embedding) == 0 && len(out.EmbeddingAlt) > 0 {
		out.Embedding = out.EmbeddingAlt
	}

	// fallback แบบโค้ดเดิม (เผื่อ colab ส่ง style_vector มาแทน)
	vec := out.StyleVectorV16
	if len(vec) == 0 && len(out.StyleVectorV16Norm) > 0 {
		vec = out.StyleVectorV16Norm
	}
	if len(vec) == 0 && len(out.StyleVector) > 0 {
		vec = out.StyleVector
	}
	out.StyleVectorV16 = vec

	if len(out.StyleVectorV16) != 16 {
		return nil, fmt.Errorf("invalid style vector v16 len=%d (want 16)", len(out.StyleVectorV16))
	}

	labelStr := "nil"
	if out.StyleLabel != nil {
		labelStr = *out.StyleLabel
	}

	log.Printf("[COLAB][EXTRACT] OK time=%s doc=%d label=%v vec_len=%d",
		time.Since(start), out.DocumentID, labelStr, len(out.StyleVectorV16))

	return &out, nil
}
