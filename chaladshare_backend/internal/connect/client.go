package connect

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client

	// timeout แยกตามงาน
	ExtractTimeout time.Duration
}

func NewFromEnv() (*Client, error) {
	base := strings.TrimRight(os.Getenv("COLAB_URL"), "/")
	if base == "" {
		return nil, fmt.Errorf("COLAB_URL is empty")
	}

	key := os.Getenv("COLAB_API_KEY")

	return &Client{
		BaseURL:        base,
		APIKey:         key,
		HTTP:           &http.Client{},
		ExtractTimeout: 180 * time.Second, // เท่าของเดิม
	}, nil
}

func (c *Client) postPDFWithField(ctx context.Context, endpoint string, documentID int, pdfPath string, fileField string) (*http.Response, error) {
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	url := c.BaseURL + endpoint

	f, err := os.Open(pdfPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	_ = w.WriteField("document_id", strconv.Itoa(documentID))
	_ = w.WriteField("file_name", filepath.Base(pdfPath))

	fw, err := w.CreateFormFile(fileField, filepath.Base(pdfPath)) // ต้องเป็น "file"
	if err != nil {
		_ = w.Close()
		return nil, err
	}
	if _, err := io.Copy(fw, f); err != nil {
		_ = w.Close()
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	// log ไว้เช็คว่ามีไฟล์จริง
	if st, _ := os.Stat(pdfPath); st != nil {
		log.Printf("[COLAB] POST %s file=%s size=%d", url, filepath.Base(pdfPath), st.Size())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ngrok-skip-browser-warning", "true")
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	return c.HTTP.Do(req)
}
