package handler

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	"github.com/stsolovey/kvant_chat/internal/utils"
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
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Only POST method is allowed", h.logger)

		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Username and password are required", h.logger)

		return
	}

	user, err := h.service.GetUserByUsername(r.Context(), username)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid credentials", h.logger)

		return
	}

	if !h.service.VerifyPassword(user.HashPassword, password) {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Invalid credentials", h.logger)

		return
	}

	tokenString, err := h.service.GenerateToken(username)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to generate token", h.logger)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	responseData := map[string]interface{}{"token": tokenString}

	utils.WriteOkResponse(w, http.StatusOK, responseData, h.logger)
}
