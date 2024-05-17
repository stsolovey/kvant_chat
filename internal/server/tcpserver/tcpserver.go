package tcpserver

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/models"
)

type Server struct {
	log       *logrus.Logger
	port      string
	rooms     map[string]*models.Room
	mutex     *sync.Mutex
	listener  net.Listener
	connUsers map[net.Conn]*models.User
}

func CreateServer(port string, logger *logrus.Logger) *Server {
	return &Server{
		log:       logger,
		port:      port,
		rooms:     make(map[string]*models.Room),
		mutex:     &sync.Mutex{},
		connUsers: make(map[net.Conn]*models.User),
	}
}

func (s *Server) getUserFromConn(conn net.Conn) *models.User {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.connUsers[conn]
}

func (s *Server) Start() {
	var err error

	s.listener, err = net.Listen("tcp", ":"+s.port)
	if err != nil {
		s.log.WithError(err).Error("Error starting TCP server:", err)

		return
	}

	defer s.listener.Close()

	s.log.Info("TCP Server listening on port", s.port)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.log.WithError(err).Error("Error accepting connection:", err)

			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	clientReader := bufio.NewReader(conn)

	for {
		message, err := clientReader.ReadString('\n')
		if err != nil {
			s.log.WithError(err).Error("Error reading from client")

			return
		}

		message = strings.TrimSpace(message)
		if message == "" {
			continue
		}

		switch {
		case strings.HasPrefix(message, "/join "):
			s.handleJoin(conn, message[len("/join "):])
		case strings.HasPrefix(message, "/msg "):
			s.handleMessage(conn, message[len("/msg "):])
		default:
			if _, err := conn.Write([]byte("Unknown command\n")); err != nil {
				s.log.WithError(err).Error("Error writing to client")

				return
			}
		}
	}
}

func (s *Server) handleJoin(conn net.Conn, roomName string) {
	s.mutex.Lock()

	defer s.mutex.Unlock()

	user := s.getUserFromConn(conn)
	if user == nil {
		errorMessage := "Error: Unable to identify user\n"

		s.log.Error("Unable to identify user from connection")

		if _, err := conn.Write([]byte(errorMessage)); err != nil {
			s.log.WithError(err).Error("Failed to send identification error to client")
		}

		return
	}

	if _, ok := s.rooms[roomName]; !ok {
		s.rooms[roomName] = &models.Room{
			Name:    roomName,
			Members: make(map[*models.User]net.Conn),
		}
	}

	room := s.rooms[roomName]
	room.Members[user] = conn

	joinMessage := fmt.Sprintf("Joined room: %s\n", roomName)

	s.log.Infof("User %s joined room: %s", user.UserName, roomName)

	if _, err := conn.Write([]byte(joinMessage)); err != nil {
		s.log.WithError(err).Error("Failed to send room join confirmation to client")
	}
}

func (s *Server) handleMessage(conn net.Conn, message string) {
	s.mutex.Lock()

	defer s.mutex.Unlock()

	user := s.getUserFromConn(conn)
	if user == nil {
		errMsg := "Error: User not identified\n"

		s.log.Error("User not identified from connection")

		if _, err := conn.Write([]byte(errMsg)); err != nil {
			s.log.WithError(err).Error("Failed to write user identification error to client")
		}

		return
	}

	for _, room := range s.rooms {
		if userConn, ok := room.Members[user]; ok { // checks if user is in the room
			for _, memberConn := range room.Members {
				if memberConn != userConn { // checkint that sender does not get own message
					messageToSend := message + "\n"
					if _, err := memberConn.Write([]byte(messageToSend)); err != nil {
						s.log.WithFields(logrus.Fields{ // failing to send a message to a room member
							"room": room.Name,
							"user": user.UserName,
						}).WithError(err).Error("Failed to send message to room member")
					} else {
						// successful message delivery
						s.log.WithFields(logrus.Fields{
							"room":    room.Name,
							"user":    user.UserName,
							"message": message,
						}).Info("Message sent to room member")
					}
				}
			}
		}
	}
}
