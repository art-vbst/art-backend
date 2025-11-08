package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GCS struct {
	bucketName  string
	tokenSource oauth2.TokenSource
}

func NewGCS(bucketName string) *GCS {
	ctx := context.Background()

	tokenSource, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/devstorage.full_control")
	if err != nil {
		return &GCS{bucketName: bucketName, tokenSource: nil}
	}

	return &GCS{bucketName: bucketName, tokenSource: tokenSource}
}

func (s *GCS) Close() {
}

func (s *GCS) GetObjectName(fileName string) string {
	base := filepath.Base(strings.TrimSpace(fileName))
	safe := sanitizeFileName(base)
	if safe == "" {
		safe = "file"
	}
	return fmt.Sprintf("uploads/%d-%s", time.Now().UnixNano(), safe)
}

func (s *GCS) GetObjectURL(objectName string) string {
	escaped := url.PathEscape(objectName)
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, escaped)
}

func (s *GCS) UploadObject(objectName, contentType string, file io.Reader) error {
	token, err := s.getAccessToken()
	if err != nil {
		return err
	}

	encodedName := url.QueryEscape(objectName)
	url := fmt.Sprintf("https://storage.googleapis.com/upload/storage/v1/b/%s/o?uploadType=media&name=%s", s.bucketName, encodedName)

	req, err := http.NewRequest("POST", url, file)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s -- %s", resp.Status, body)
	}

	return nil
}

func (s *GCS) DeleteObject(objectName string) error {
	token, err := s.getAccessToken()
	if err != nil {
		return err
	}

	encodedName := url.QueryEscape(objectName)
	url := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o/%s", s.bucketName, encodedName)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %s -- %s", resp.Status, body)
	}

	return nil
}

func (s *GCS) getAccessToken() (string, error) {
	switch {
	case s.tokenSource != nil:
		return s.getAccessTokenFromDefaultCredentials()
	default:
		return s.getAccessTokenFromInternalAPI()
	}
}

func (s *GCS) getAccessTokenFromDefaultCredentials() (string, error) {
	token, err := s.tokenSource.Token()
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func (s *GCS) getAccessTokenFromInternalAPI() (string, error) {
	req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get token failed: %s -- %s", resp.Status, body)
	}

	var data struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.AccessToken, nil
}

func sanitizeFileName(name string) string {
	const maxLen = 150
	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.' || r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
		if b.Len() >= maxLen {
			break
		}
	}
	return strings.Trim(b.String(), "._")
}
