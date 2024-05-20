package main

import (
	"bufio"
	"fmt"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/config"
	"github.com/stsolovey/kvant_chat/internal/logger"
)

func main() {
	log := logger.New()

	cfg, err := config.NewClientConfig(log, "./.env")
	if err != nil {
		log.WithError(err).Panic("Failed to initialize config")
	}

	log.Infoln("Attempting to connect to server...")
	conn, err := net.Dial("tcp", cfg.ServerAddress)
	if err != nil {
		log.WithError(err).Panic("Error connecting to server")
	}

	defer func() {
		if err = conn.Close(); err != nil {
			log.WithError(err).Panic("Error closing connecting to server")
		}
	}()

	log.Infoln("Connected to server successfully.")

	handleConnection(log, conn)
}

func handleConnection(log *logrus.Logger, conn net.Conn) {
	log.Infoln("Enter authentication token:")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		token := scanner.Text()
		_, err := conn.Write([]byte(token + "\n"))
		if err != nil {
			fmt.Printf("Failed to send token: %v\n", err)
			return
		}
	}
	go receiveMessages(log, conn)

	for {
		if scanner.Scan() {
			message := scanner.Text()
			_, err := conn.Write([]byte(message + "\n"))
			if err != nil {
				fmt.Printf("Failed to send message: %v\n", err)
				continue
			}
		}
	}
}

func receiveMessages(log *logrus.Logger, conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		log.Infoln("Received:", message)
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error receiving messages: %v\n", err)
	}
}
