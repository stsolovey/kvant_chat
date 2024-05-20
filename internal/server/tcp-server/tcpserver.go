package tcpserver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt"
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
	return &Server{
		cfg:         config,
		log:         logger,
		rooms:       make(map[string]*models.Room),
		mutex:       &sync.Mutex{},
		connUsers:   make(map[net.Conn]*models.User),
		authService: authService,
	}
}

func (s *Server) Start(ctx context.Context) error {
	var err error

	s.listener, err = net.Listen("tcp", ":"+s.cfg.TCPPort)
	if err != nil {
		return fmt.Errorf("error starting TCP server: %w", err)
	}

	defer func() {
		if err = s.listener.Close(); err != nil {
			s.log.Errorf("err=s.listener.Close(): %v", err)
		}
	}()

	s.log.Info("TCP Server listening on port", s.cfg.TCPPort)

	connChan := make(chan net.Conn)
	errChan := make(chan error)

	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				errChan <- err

				return
			}
			connChan <- conn
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.log.Info("TCP server shutdown initiated.")

				return
			case err := <-errChan:
				if !errors.Is(err, net.ErrClosed) {
					s.log.WithError(err).Error("Error accepting connection")
				}

				continue
			case conn := <-connChan:
				go s.handleConnection(conn)
			}
		}
	}()

	<-ctx.Done()

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			s.log.WithError(err).Panic("Error closing connecting")
		}
	}()

	clientReader := bufio.NewReader(conn)

	token, err := clientReader.ReadString('\n')
	if err != nil {
		s.log.WithError(err).Error("Failed to read the authentication token")
		conn.Write([]byte("Failed to read token\n"))

		return
	}

	token = strings.TrimSpace(token)

	jwtToken, err := s.authService.ValidateToken(token)
	if err != nil {
		s.log.WithError(err).Error("Authentication failed")
		conn.Write([]byte("Authentication failed\n"))

		return
	}

	username := jwtToken.Claims.(jwt.MapClaims)["username"].(string)

	user, err := s.authService.GetUserByUsername(context.Background(), username)
	if err != nil {
		s.log.WithError(err).Error("Failed to retrieve user data")
		conn.Write([]byte("Failed to retrieve user data\n"))

		return
	}

	s.log.Infof("User %s authenticated successfully", user.UserName)
	s.mutex.Lock()
	s.connUsers[conn] = user
	s.joinRoom("general", user, conn)
	s.mutex.Unlock()
	conn.Write([]byte("Welcome to the chat server!\n"))

	for {
		message, err := clientReader.ReadString('\n')
		if err != nil {
			s.log.WithError(err).Error("Error reading from client")
			s.leaveRoom("general", user)

			return
		}

		s.processCommand(user, strings.TrimSpace(message), conn)
	}
}

func (s *Server) processCommand(user *models.User, command string, conn net.Conn) {
	switch {
	case strings.HasPrefix(command, "/join "):
		roomName := command[len("/join "):]
		s.joinRoom(roomName, user, conn)
	case strings.HasPrefix(command, "/leave "):
		roomName := command[len("/leave "):]
		s.leaveRoom(roomName, user)
	case strings.HasPrefix(command, "/msg "):
		messageContent := command[len("/msg "):]
		s.handleMessage(user, messageContent, conn)
	default:
		conn.Write([]byte("Unknown command\n"))
	}
}

func (s *Server) handleMessage(user *models.User, message string, conn net.Conn) {
	s.mutex.Lock()

	defer s.mutex.Unlock()

	if conn == nil {
		s.log.Warn("Connection is nil, which is unexpected")
	}

	room, exists := s.rooms["general"]
	if !exists {
		s.log.Error("General room does not exist")

		return
	}

	for member, memberConn := range room.Members {
		if member != user { // exclude sender
			messageToSend := fmt.Sprintf("%s: %s\n", user.UserName, message)
			if _, err := memberConn.Write([]byte(messageToSend)); err != nil {
				s.log.WithFields(logrus.Fields{
					"room": room.Name,
					"user": user.UserName,
				}).WithError(err).Error("Failed to send message to room member")
			}
		}
	}
}

func (s *Server) joinRoom(roomName string, user *models.User, conn net.Conn) {
	if _, exists := s.rooms[roomName]; !exists {
		s.rooms[roomName] = &models.Room{
			Name:    roomName,
			Members: make(map[*models.User]net.Conn),
		}
	}

	s.rooms[roomName].Members[user] = conn
	s.log.Infof("User %s joined the room %s", user.UserName, roomName)
}

func (s *Server) leaveRoom(roomName string, user *models.User) {
	if room, exists := s.rooms[roomName]; exists {
		delete(room.Members, user)
		s.log.Infof("User %s left the room %s", user.UserName, roomName)
	}
}
