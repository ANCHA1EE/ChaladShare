package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type AISummaryHandler struct {
	colabURL string
	apiKey   string
	client   *http.Client
}

func NewAISummaryHandler() *AISummaryHandler {
	return &AISummaryHandler{
		colabURL: strings.TrimSpace(os.Getenv("COLAB_URL")),
		apiKey:   strings.TrimSpace(os.Getenv("COLAB_API_KEY")),
		client: &http.Client{
			Timeout: 10 * time.Minute, // กันสรุปนาน
		},
	}
}

func (h *AISummaryHandler) Summarize(c *gin.Context) {
	if h.colabURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "COLAB URL is not set"})
		return
	}

	// sure to point at /summarize
	u, err := url.Parse(h.colabURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "COLAB_URL is invalid"})
		return
	}

	// ถ้า COLAB_URL เป็น base (path ว่างหรือ "/") ให้เซ็ตเป็น /summarize
	if u.Path == "" || u.Path == "/" {
		u.Path = "/summarize"
	} else if !strings.HasSuffix(strings.TrimRight(u.Path, "/"), "/summarize") {
		u.Path = strings.TrimRight(u.Path, "/") + "/summarize"
	}
	targetURL := u.String()

	// รับไฟล์จาก React: form-data key = "file"
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file (key: file)"})
		return
	}

	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open uploaded file"})
		return
	}
	defer src.Close()

	// สร้าง multipart ส่งต่อไป Colab
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", fh.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create multipart"})
		return
	}
	if _, err := io.Copy(part, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot write multipart"})
		return
	}
	if err := writer.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot close multipart"})
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, targetURL, &buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create request"})
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("ngrok-skip-browser-warning", "true")

	if h.apiKey != "" {
		req.Header.Set("X-API-Key", h.apiKey)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to call colab", "detail": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// forward status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.Data(resp.StatusCode, "application/json; charset=utf-8", body)
		return
	}

	var js map[string]any
	if err := json.Unmarshal(body, &js); err != nil {
		c.Data(200, "text/plain; charset=utf-8", body)
		return
	}
	c.JSON(http.StatusOK, js)
}
