package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var errAssertingClaimsJWT = errors.New("error asserting claims to jwt.MapClaims")

func generateToken(username string, mySigningKey []byte) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errAssertingClaimsJWT
	}

	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", fmt.Errorf("generateToken failed: %w", err)
	}

	return tokenString, nil
}

func makeServer(log *logrus.Logger) *http.Server {
	server := &http.Server{
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Starting server at :8080")

	return server
}

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	err := godotenv.Load()
	if err != nil {
		log.WithError(err).Panic("Error loading .env file")
	}

	mySigningKey := []byte(os.Getenv("JWT_SECRET"))
	if len(mySigningKey) == 0 {
		log.WithError(err).Panic("JWT_SECRET is not set")
	}

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)

			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)

			return
		}

		username := r.FormValue("username")
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)

			return
		}

		tokenString, err := generateToken(username, mySigningKey)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)

			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:  "session_token",
			Value: tokenString,
			Path:  "/",
		})

		if _, err := w.Write([]byte("Logged in successfully with token: " + tokenString)); err != nil {
			log.WithError(err).Panic("Failed to write response")
		}
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	server := makeServer(log)

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Panic("Server stopped unexpectedly")
		}
	}()

	<-ctx.Done()

	if err := server.ListenAndServe(); err != nil {
		log.WithError(err).Panic("Error starting server: %w\n", err)
	}
}
