package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	errUsernameEmpty = errors.New("username cannot be empty")
	errStatusNotOK   = errors.New("wrong status")
)

func main() {
	log := logrus.New()

	os.Stdout.WriteString("Please enter your username: ")

	var username string

	fmt.Scanln(&username)

	if username == "" {
		log.WithError(errUsernameEmpty).Panic("Username cannot be empty.")
	}

	url := "http://localhost:8080/login"
	data := "username=" + username

	const timeoutDuration = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBufferString(data))
	if err != nil {
		log.WithError(err).Panic("Error creating request")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Panic("Error sending request.")
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Panic("Error reading response.")
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Failed to log in: %s\n", body)
		log.WithError(errStatusNotOK).Panic("Failed to log in.")
	}

	log.Infof("Response from server: %s", string(body))
}
