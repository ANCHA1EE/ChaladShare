package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type StorageClient interface {
	UploadLocalFile(ctx context.Context, objectPath string, localPath string) (publicURL string, err error)
	Delete(ctx context.Context, objectPath string) error
	ObjectPathFromPublicURL(publicURL string) (objectPath string, ok bool)
}

type SupabaseStorage struct {
	baseURL    string
	serviceKey string
	bucket     string
	httpClient *http.Client
}

func NewSupabaseStorageFromEnv() (*SupabaseStorage, error) {
	baseURL := strings.TrimRight(os.Getenv("SUPABASE_URL"), "/")
	key := strings.TrimSpace(os.Getenv("SUPABASE_SERVICE_ROLE_KEY"))
	if key == "" {
		key = strings.TrimSpace(os.Getenv("SUPABASE_ANON_KEY"))
	}
	bucket := strings.TrimSpace(os.Getenv("SUPABASE_STORAGE_BUCKET"))

	if baseURL == "" || key == "" || bucket == "" {
		return nil, errors.New("missing env: SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY(or SUPABASE_ANON_KEY), SUPABASE_STORAGE_BUCKET")
	}

	return &SupabaseStorage{
		baseURL:    baseURL,
		serviceKey: key,
		bucket:     bucket,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (s *SupabaseStorage) UploadLocalFile(ctx context.Context, objectPath string, localPath string) (string, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// PUT
	u := fmt.Sprintf("%s/storage/v1/object/%s/%s?upsert=true",
		s.baseURL,
		url.PathEscape(s.bucket),
		escapeObjectPath(objectPath),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, f)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}

	ct := mime.TypeByExtension(strings.ToLower(filepath.Ext(localPath)))
	if ct == "" {
		ct = "application/octet-stream"
	}

	req.Header.Set("Content-Type", ct)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("apikey", s.serviceKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase upload failed: %s - %s", resp.Status, string(b))
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		s.baseURL,
		s.bucket,
		escapeObjectPath(objectPath),
	)

	return publicURL, nil
}

func (s *SupabaseStorage) Delete(ctx context.Context, objectPath string) error {
	u := fmt.Sprintf("%s/storage/v1/object/%s/%s",
		s.baseURL,
		url.PathEscape(s.bucket),
		escapeObjectPath(objectPath),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("apikey", s.serviceKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase delete failed: %s - %s", resp.Status, string(b))
	}
	return nil
}

func (s *SupabaseStorage) ObjectPathFromPublicURL(publicURL string) (string, bool) {
	prefix := fmt.Sprintf("%s/storage/v1/object/public/%s/", strings.TrimRight(s.baseURL, "/"), s.bucket)
	if !strings.HasPrefix(publicURL, prefix) {
		return "", false
	}
	suffix := strings.TrimPrefix(publicURL, prefix)

	decoded, err := url.PathUnescape(suffix)
	if err == nil && decoded != "" {
		suffix = decoded
	}
	return suffix, true
}

func escapeObjectPath(p string) string {
	parts := strings.Split(p, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}
