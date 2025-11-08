package router

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/art-vbst/art-backend/internal/platform/config"
	"github.com/art-vbst/art-backend/internal/platform/storage"
)

func serveLocalStorage(localStorage *storage.LocalStorage) http.Handler {
	storageDir := localStorage.GetStorageDir()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !config.IsDebug() {
			http.NotFound(w, r)
			return
		}

		filePath := filepath.Join(storageDir, r.URL.Path)

		ext := filepath.Ext(filePath)
		if contentType := getContentType(ext); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}

		http.ServeFile(w, r, filePath)
	})
}

func getContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		return ""
	}
}
