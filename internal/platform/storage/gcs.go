package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/art-vbst/art-backend/internal/platform/config"
)

type GCS struct {
	bucketName  string
	accessToken string
}

func NewGCS(env *config.Config) *GCS {
	return &GCS{bucketName: env.GCSBucketName, accessToken: env.GCSAccessToken}
}

func (s *GCS) Close() {
}

func (s *GCS) GetObjectName(fileName string) string {
	return fmt.Sprintf("uploads/%d-%s", time.Now().UnixNano(), fileName)
}

func (s *GCS) GetObjectURL(objectName string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, objectName)
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
	if s.accessToken != "" {
		return s.accessToken, nil
	}

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
