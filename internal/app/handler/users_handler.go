package handler

import (
	// "encoding/json"
	// "fmt"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	"github.com/stsolovey/kvant_chat/internal/models"
	"golang.org/x/crypto/bcrypt"
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

	username := r.FormValue("username")
	password := r.FormValue("password")

	if len(username) < 6 {
		http.Error(w, "Username must be at least 6 characters long", http.StatusBadRequest)
		return
	}

	if len(password) < 6 {
		http.Error(w, "Password must be at least 6 characters long", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	userCreateInput := models.UserCreateInput{
		Name:         username,
		HashPassword: string(hashedPassword),
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
	w.Write(response)
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
