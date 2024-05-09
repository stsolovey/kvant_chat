package handler

import (
	// "encoding/json"
	// "fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	// "github.com/stsolovey/kvant_chat/internal/models"
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
