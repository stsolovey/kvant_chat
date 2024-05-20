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

type UsersHandler struct {
	service service.UsersServiceInterface
	logger  *logrus.Logger
}

func NewUsersHandler(s service.UsersServiceInterface, logger *logrus.Logger) *UsersHandler {
	return &UsersHandler{
		service: s,
		logger:  logger,
	}
}

func (h *UsersHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErrorResponse(w, http.StatusMethodNotAllowed, "Only POST method is allowed", h.logger)

		return
	}

	var input models.UserRegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid JSON data", h.logger)

		return
	}

	userResponse, token, err := h.service.RegisterUser(r.Context(), input)
	if err != nil {
		handleRegisterServiceError(w, err, h.logger)

		return
	}

	responseData := map[string]interface{}{"user": userResponse, "token": token}
	utils.WriteOkResponse(w, http.StatusCreated, responseData, h.logger)
}

func handleRegisterServiceError(w http.ResponseWriter, err error, log *logrus.Logger) {
	var statusCode int

	var errMsg string

	switch {
	case errors.Is(err, models.ErrUsernameExists):
		statusCode = http.StatusConflict
		errMsg = "Username already exists"
	case errors.Is(err, models.ErrUsernameTooShort), errors.Is(err, models.ErrPasswordTooShort):
		statusCode = http.StatusBadRequest
		errMsg = err.Error()
	default:
		statusCode = http.StatusInternalServerError
		errMsg = "Internal server error"
	}

	utils.WriteErrorResponse(w, statusCode, errMsg, log)
}
