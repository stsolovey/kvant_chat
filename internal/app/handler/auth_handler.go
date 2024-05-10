package handler

import (
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
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	tokenString, err := h.service.GenerateToken(username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: tokenString,
		Path:  "/",
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged in successfully with token: " + tokenString))
}
