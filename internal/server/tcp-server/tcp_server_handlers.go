package tcpserver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stsolovey/kvant_chat/internal/models"
)

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) error {
	defer func() {
		if err := conn.Close(); err != nil {
			s.log.WithError(err).Panic("Error closing connecting")
		}
	}()

	user, err := s.getUserFromConn(ctx, conn)
	if err != nil {
		if _, err = conn.Write([]byte("Failed to retrieve user data\n")); err != nil {
			return fmt.Errorf("handleConnection(...) getUserFromConn(...) conn.Write: %w", err)
		}

		return fmt.Errorf("handleConnection(...) getUserFromConn(...): %w", err)
	}

	user.Conn = conn

	s.log.Infof("User %s authenticated successfully", user.UserName)

	s.mutex.Lock()
	s.connUsers[conn] = user
	s.rooms["general"].Members[user] = conn
	s.mutex.Unlock()

	content := "Welcome to the chat, " + user.UserName + "!"

	if err := s.sendMessage("Server", user.UserName, content, conn); err != nil {
		return fmt.Errorf("handleConnection(...) s.sendMessage(...): %w", err)
	}

	joinMsg := models.Message{
		// Receiver:  "everyone",.
		Content:   user.UserName + " has joined the chat!",
		Sender:    "Server",
		CreatedAt: time.Now(),
	}

	if err := s.broadcastToAll(joinMsg, nil); err != nil {
		return fmt.Errorf("handleConnection(...) s.broadcastToAll(...): %w", err)
	}

	if err := s.handleMessages(conn, user); err != nil {
		return fmt.Errorf("handleConnection(...) s.handleMessages(...): %w", err)
	}

	return nil
}

func (s *Server) handleMessages(conn net.Conn, user *models.User) error {
	reader := bufio.NewReader(conn)

	for {
		msgText, err := reader.ReadString('\n')
		if err != nil {
			s.leaveRoom("general", user)

			return fmt.Errorf("handleMessages reader.ReadString(...): %w", err)
		}

		var msg models.Message
		if err := json.Unmarshal([]byte(msgText), &msg); err != nil {
			fmt.Fprintf(conn, "Error parsing message: %v\n", err)

			continue
		}

		if msg.Receiver != "" && msg.Receiver != "everyone" {
			if err := s.handleDirectMessage(msg, user); err != nil {
				return fmt.Errorf("handleMessages s.handleDirectMessage(...): %w", err)
			}
		} else {
			if err := s.broadcastToRoom("general", msg, user); err != nil {
				return fmt.Errorf("handleMessages s.broadcastToRoom(...): %w", err)
			}
		}
	}
}

func (s *Server) handleDirectMessage(msg models.Message, sender *models.User) error {
	recipientName := msg.Receiver
	found := false

	for _, room := range s.rooms {
		for user, conn := range room.Members {
			if user.UserName == recipientName {
				if err := s.sendMessage(sender.UserName, recipientName, msg.Content, conn); err != nil {
					return fmt.Errorf("handleDirectMessage failed to send message when found user: %w", err)
				}

				found = true

				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		errorMsg := fmt.Sprintf("User %s not found.", recipientName)
		if err := s.sendMessage("Server", sender.UserName, errorMsg, sender.Conn); err != nil {
			return fmt.Errorf("handleDirectMessage if !found s.sendMessage(...): %w", err)
		}
	}

	return nil
}

func (s *Server) sendMessage(
	sender string,
	receiver string,
	content string,
	conn net.Conn,
) error {
	message := models.Message{
		Sender:    sender,
		Receiver:  receiver,
		Content:   content,
		CreatedAt: time.Now(),
	}

	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("sendMessage(...) failed to marshal message: %w", err)
	}

	if _, err = conn.Write(append(jsonMsg, '\n')); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (s *Server) getUserFromConn(ctx context.Context, conn net.Conn) (*models.User, error) {
	clientReader := bufio.NewReader(conn)

	token, err := clientReader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("getUserFromToken clientReader.ReadString: %w", err)
	}

	token = strings.TrimSpace(token)

	jwtToken, err := s.authService.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("getUserFromToken: %w", models.ErrParseJWTClaimsAsMapClaims)
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, fmt.Errorf("username claim is not a string: %w", models.ErrUsernameClaimIsNotString)
	}

	user, err := s.authService.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user data: %w", err)
	}

	return user, nil
}
