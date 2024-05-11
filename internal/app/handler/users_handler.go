package handler

import (
	// "encoding/json"
	// "fmt"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	"github.com/stsolovey/kvant_chat/internal/models"
)

type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Error *string     `json:"error,omitempty"`
}

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
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)

		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)

		return
	}

	userCreateInput := models.UserCreateInput{
		Name:         r.FormValue("username"),
		HashPassword: r.FormValue("password"),
	}

	createdUser, err := h.service.CreateUser(r.Context(), userCreateInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	response, err := json.Marshal(createdUser)
	if err != nil {
		http.Error(w, "Failed to parse user response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		h.logger.Error("Failed to write response: ", err)
	}
}

func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
}

func (h *UsersHandler) GetUser(w http.ResponseWriter, r *http.Request) {
}

func (h *UsersHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
}

func (h *UsersHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
}

func (h *UsersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
}
