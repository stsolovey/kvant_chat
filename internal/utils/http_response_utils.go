package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/models"
)

func WriteOkResponse(w http.ResponseWriter, statusCode int, data any, log *logrus.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(models.HTTPResponse{Data: data})
	if err != nil {
		log.WithError(err).Warn("Failed to encode response")
	}
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, description string, log *logrus.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(models.HTTPResponse{Error: description})
	if err != nil {
		log.WithError(err).Warn("Failed to encode response")
	}
}
