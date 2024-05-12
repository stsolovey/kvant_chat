package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/app/service"
)

type AuthHandler struct {
	service service.AuthServiceInterface
	logger  *logrus.Logger
}

func NewAuthHandler(s service.AuthServiceInterface, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		service: s,
		logger:  logger,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)

		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)

		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUserByUsername(r.Context(), username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)

		return
	}

	if !h.service.VerifyPassword(user.HashPassword, password) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)

		return
	}

	tokenString, err := h.service.GenerateToken(username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"token": tokenString}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error generating response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

	_, err = w.Write(jsonResponse)
	if err != nil {
		h.logger.Error("Failed to write response: ", err)
	}
}
