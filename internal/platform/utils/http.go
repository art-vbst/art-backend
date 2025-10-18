package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func RespondError(w http.ResponseWriter, status int, message string) {
	log.Printf("%d %s\n", status, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func RespondServerError(w http.ResponseWriter) {
	RespondError(w, http.StatusInternalServerError, "an unknown error occurred")
}
