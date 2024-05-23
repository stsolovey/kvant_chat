package tcpserver

import (
	"encoding/json"
	"fmt"

	"github.com/stsolovey/kvant_chat/internal/models"
)

func (s *Server) broadcastToRoom(roomName string, msg models.Message, exceptUser *models.User) error {
	msg.Receiver = ""

	s.mutex.Lock()
	defer s.mutex.Unlock()

	room, exists := s.rooms[roomName]
	if !exists {
		return fmt.Errorf("broadcastToRoom, s.rooms[roomName]: %w", models.ErrRoomNotExists)
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("broadcastToRoom(...) failed to marshal message: %w", err)
	}

	for user, conn := range room.Members {
		if user != exceptUser {
			_, err := conn.Write(append(jsonData, '\n'))
			if err != nil {
				return fmt.Errorf("broadcastToRoom username - %s, room - %s: %w",
					user.UserName, roomName, err)
			}
		}
	}

	return nil
}

func (s *Server) broadcastToAll(msg models.Message, exceptUser *models.User) error {
	msg.Receiver = ""

	s.mutex.Lock()
	defer s.mutex.Unlock()

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("broadcastToAll(...) failed to marshal message: %w", err)
	}

	for _, room := range s.rooms {
		for user, conn := range room.Members {
			if user != exceptUser {
				_, err := conn.Write(append(jsonData, '\n'))
				if err != nil {
					return fmt.Errorf("broadcastToRoom username - %s, room - %s: %w",
						user.UserName, room.Name, err)
				}
			}
		}
	}

	return nil
}
