package connect

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	recmodels "chaladshare_backend/internal/recommend/models"
)

func (c *Client) RecommendFromLiked(req recmodels.ColabRecommendFromLikedReq) (*recmodels.ColabRecommendFromLikedResp, error) {
	if c == nil {
		return nil, fmt.Errorf("connect client is nil")
	}
	base := strings.TrimRight(c.BaseURL, "/")
	if base == "" {
		return nil, fmt.Errorf("COLAB_URL is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal recommend req: %w", err)
	}

	url := base + "/recommend/from-liked"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("ngrok-skip-browser-warning", "true")
	if c.APIKey != "" {
		httpReq.Header.Set("X-API-Key", c.APIKey)
	}

	client := c.HTTP
	if client == nil {
		client = &http.Client{}
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call colab recommend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("recommend status %d: %s", resp.StatusCode, string(body))
	}

	var out recmodels.ColabRecommendFromLikedResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode recommend resp: %w", err)
	}
	return &out, nil
}
