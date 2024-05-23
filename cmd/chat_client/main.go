package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/config"
	"github.com/stsolovey/kvant_chat/internal/logger"
	"github.com/stsolovey/kvant_chat/internal/models"
)

func main() {
	log := logger.NewTextFormat()

	cfg, err := config.NewClientConfig(log, "./.env")
	if err != nil {
		log.WithError(err).Panic("Failed to initialize config")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn := connectToTCPServer(cfg.TCPServerAddr, log)

	defer func() {
		if err = conn.Close(); err != nil {
			log.WithError(err).Panic("Error closing connecting to server")
		}
	}()

	var wg sync.WaitGroup

	wg.Add(1) // wait for the initial message

	go receiveMessages(conn, log, &wg)

	username, token := authenticateUser(ctx, bufio.NewReader(os.Stdin), cfg, log)
	authenticateWithTCPServer(token, conn, log)

	wg.Wait() // wait here until the initial message is received

	log.Println("Enter messages to send to the chat server:")
	log.Println("Type '" + color.GreenString("/logout") + "' to disconnect and log out.")
	log.Println("Type '" + color.CyanString("@username ") + color.BlueString("your_message") +
		"' to send a direct message to 'username'.")

	sendMessages(ctx, cancel, conn, bufio.NewReader(os.Stdin), log, username)
}

func authenticateUser(
	ctx context.Context,
	reader *bufio.Reader,
	cfg *config.Config,
	log *logrus.Logger,
) (string, string) {
	for {
		fmt.Println("Choose an option:") //nolint:forbidigo
		fmt.Println("1: Register")       //nolint:forbidigo
		fmt.Println("2: Login")          //nolint:forbidigo

		fmt.Print("Option: ") //nolint:forbidigo

		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		var url string

		var creds models.Credentials

		fmt.Println("Enter username:") //nolint:forbidigo
		fmt.Print("Username: ")        //nolint:forbidigo

		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		fmt.Println("Enter password:") //nolint:forbidigo
		fmt.Print("Password: ")        //nolint:forbidigo

		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		switch option {
		case "1":
			url = cfg.RegisterURL
		case "2":
			url = cfg.LoginURL
		default:
			log.Print("Invalid option. Please choose 1 for Register, 2 for Login, or 3 to Exit.")

			continue
		}

		creds = models.Credentials{Username: username, Password: password}

		token, err := sendRequest(ctx, log, url, creds)
		if err != nil {
			log.Error("Error during authentication: ", err)

			continue // prompt again after error
		}

		log.Info("Authentication successful. Token received.")

		return username, token
	}
}

func sendRequest(ctx context.Context, log *logrus.Logger, url string, data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("chat_client sendRequest json.Marshal(data): %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("chat_client sendRequest http.NewRequestWithContext(...): %w", err)
	}

	const timeoutDuration = time.Second * 10

	c := http.Client{
		Timeout: timeoutDuration,
	}

	resp, err := c.Do(req)
	if err != nil {
		return "", fmt.Errorf("chat_client sendRequest c.Do(...): %w", err)
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.WithError(err).Panic("Error closing connecting to server")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("chat_client sendRequest io.ReadAll(...): %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return "", fmt.Errorf("chat_client sendRequest json.Unmarshal(...): %w", err)
		}

		return "", fmt.Errorf("server error: %s, %w", errResp.Error, models.ErrWrongStatusCode)
	}

	var response struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("chat_client sendRequest json.Unmarshal(...): %w", err)
	}

	return response.Data.Token, nil
}

func connectToTCPServer(serverAddr string, log *logrus.Logger) net.Conn {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.WithError(err).Panic("Error connecting to TCP server")
	}

	log.Info("Connected to TCP server successfully.")

	return conn
}

func authenticateWithTCPServer(token string, conn net.Conn, log *logrus.Logger) {
	_, err := conn.Write([]byte(token + "\n"))
	if err != nil {
		log.WithError(err).Panic("Failed to send authentication token to TCP server")
	}

	log.Info("Authentication token sent.")
}

func receiveMessages(conn net.Conn, log *logrus.Logger, wg *sync.WaitGroup) {
	scanner := bufio.NewScanner(conn)
	isFirstMessage := true

	for scanner.Scan() {
		jsonInput := scanner.Text()

		var msg models.Message

		err := json.Unmarshal([]byte(jsonInput), &msg)
		if err != nil {
			log.WithError(err).Error("Failed to parse message")

			continue
		}

		formattedTime := msg.CreatedAt.Format("2006-01-02 15:04:05")

		var recipient string
		if msg.Receiver == "" {
			recipient = "everyone"
		} else {
			recipient = msg.Receiver
		}

		messagePrefix := color.GreenString("MESSAGE")
		formattedMessage := fmt.Sprintf(
			"%s[%s] %s to %s: %s",
			messagePrefix,
			formattedTime,
			msg.Sender,
			recipient,
			msg.Content)

		if isFirstMessage {
			fmt.Println(formattedMessage) //nolint:forbidigo

			isFirstMessage = false

			wg.Done() // signal that first message received
		} else {
			fmt.Println(formattedMessage) //nolint:forbidigo
		}
	}

	if err := scanner.Err(); err != nil {
		log.WithError(err).Error("Error receiving messages from server")
	}
}

func sendMessages(ctx context.Context, cancel context.CancelFunc, //nolint:funlen
	conn net.Conn, reader *bufio.Reader, log *logrus.Logger, username string,
) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			log.Print("Enter command or message:\n")

			input, err := reader.ReadString('\n')
			if err != nil {
				log.WithError(err).Error("Failed to read input")

				continue
			}

			input = strings.TrimSpace(input)

			if input == "/logout" {
				cancel()

				return
			}

			if input == "" {
				continue
			}

			currentTime := time.Now()
			formattedTime := currentTime.Format("2006-01-02 15:04:05")
			recipient := "everyone"

			const receiverSeparator = 2

			if strings.HasPrefix(input, "@") {
				parts := strings.SplitN(input, " ", receiverSeparator)
				if len(parts) > 1 {
					recipient = parts[0][1:] // removing '@' from username
					input = parts[1]
				}
			}

			messagePrefix := color.GreenString("MESSAGE")
			ouptutMessage := fmt.Sprintf("%s[%s] You to %s: %s\n",
				messagePrefix, formattedTime, recipient, input)
			fmt.Print(ouptutMessage) //nolint:forbidigo

			msg := models.Message{
				Sender: username, Receiver: recipient,
				Content: input, CreatedAt: currentTime,
			}

			jsonData, err := json.Marshal(msg)
			if err != nil {
				log.WithError(err).Error("Failed to serialize message")

				continue
			}

			_, err = conn.Write(append(jsonData, '\n'))
			if err != nil {
				log.WithError(err).Error("Failed to send message to server")

				continue
			}
		}
	}
}
