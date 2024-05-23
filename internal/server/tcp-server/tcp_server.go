package tcpserver

import (
	"net"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	"github.com/stsolovey/kvant_chat/internal/config"
	"github.com/stsolovey/kvant_chat/internal/models"
)

type Server struct {
	cfg         *config.Config
	log         *logrus.Logger
	rooms       map[string]*models.Room
	mutex       *sync.Mutex
	listener    net.Listener
	connUsers   map[net.Conn]*models.User
	authService *service.AuthService
}

func CreateServer(
	config *config.Config,
	logger *logrus.Logger,
	authService *service.AuthService,
) *Server {
	rooms := make(map[string]*models.Room)
	rooms["general"] = &models.Room{Name: "general", Members: make(map[*models.User]net.Conn)}

	return &Server{
		cfg:         config,
		log:         logger,
		rooms:       rooms,
		mutex:       &sync.Mutex{},
		connUsers:   make(map[net.Conn]*models.User),
		authService: authService,
	}
}

func (s *Server) leaveRoom(roomName string, user *models.User) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if room, exists := s.rooms[roomName]; exists {
		delete(room.Members, user)
	}
}
