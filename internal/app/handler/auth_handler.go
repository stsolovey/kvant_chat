package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	"github.com/stsolovey/kvant_chat/internal/models"
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

	var loginRequest models.UserLoginInput
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON data", h.logger)

		return
	}

	token, err := h.service.LoginUser(r.Context(), loginRequest)
	if err != nil {
		handleLoginServiceError(w, err, h.logger)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	responseData := map[string]interface{}{"token": token}
	utils.WriteOkResponse(w, http.StatusOK, responseData, h.logger)
}

func handleLoginServiceError(w http.ResponseWriter, err error, log *logrus.Logger) {
	var statusCode int

	var errMsg string

	switch {
	case errors.Is(err, models.ErrCredentialsRequired):
		statusCode = http.StatusUnauthorized
		errMsg = err.Error()
	case errors.Is(err, models.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized
		errMsg = err.Error()
	default:
		statusCode = http.StatusInternalServerError
		errMsg = "Internal server error"
	}

	utils.WriteErrorResponse(w, statusCode, errMsg, log)
}
