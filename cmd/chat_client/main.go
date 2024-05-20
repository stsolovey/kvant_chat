package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/stsolovey/kvant_chat/internal/logger"
	"github.com/stsolovey/kvant_chat/internal/models"
)

func main() {
	log := logger.NewTextFormat()

	reader := bufio.NewReader(os.Stdin)
	serverURL := "http://localhost:8080/api/v1/user"

	log.Info("Choose an option:")
	log.Info("1: Register")
	log.Info("2: Login")

	fmt.Print("Option: ")
	option, _ := reader.ReadString('\n')

	log.Info("Enter username:")
	fmt.Print("Username: ")
	username, _ := reader.ReadString('\n')
	username = username[:len(username)-1]

	log.Info("Enter password:")
	fmt.Print("Password: ")
	password, _ := reader.ReadString('\n')
	password = password[:len(password)-1]

	switch option[0] {
	case '1':
		log.Info("Registering...")
		creds := models.UserRegisterInput{UserName: username, HashPassword: password}
		url := serverURL + "/register"
		response, err := sendRequest(url, creds)
		if err != nil {
			log.Error("Registration error: ", err)
			return
		}
		log.Info("Registration response: ", response)
	case '2':
		log.Info("Logging in...")
		creds := models.UserLoginInput{UserName: username, Password: password}
		url := serverURL + "/login"
		response, err := sendRequest(url, creds)
		if err != nil {
			log.Error("Login error: ", err)
			return
		}
		log.Info("Login response: ", response)
	default:
		log.Info("Invalid option")
	}
}

func sendRequest(url string, data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return "", err // Error unmarshalling the error response
		}
		return "", fmt.Errorf("server error: %s", errResp.Error)
	}

	var response struct {
		Data struct {
			Token string               `json:"token"`
			User  *models.UserResponse `json:"user,omitempty"`
		} `json:"data"`
		Token string `json:"token"` // For login response
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err // Error unmarshalling the data response
	}

	// Displaying user data if available (for register response)
	userDisplay := ""
	if response.Data.User != nil {
		userDisplay = fmt.Sprintf(", User: %s, CreatedAt: %s, UpdatedAt: %s",
			response.Data.User.UserName, response.Data.User.CreatedAt, response.Data.User.UpdatedAt)
	}

	return fmt.Sprintf("Token: %s%s", response.Data.Token, userDisplay), nil
}
