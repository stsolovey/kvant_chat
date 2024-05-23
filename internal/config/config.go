package config

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var (
	errMissingHost      = errors.New("postgresHost environment variable is missing")
	errMissingPort      = errors.New("postgresPort environment variable is missing")
	errMissingUser      = errors.New("postgresUser environment variable is missing")
	errMissingPassword  = errors.New("postgresPassword environment variable is missing")
	errMissingDB        = errors.New("postgresDB environment variable is missing")
	errMissingAppPort   = errors.New("appPort environment variable is missing")
	errMissingTCPPort   = errors.New("tcpPort environment variable is missing")
	errMissingJwtSecret = errors.New("jwtSecret environment variable is missing")

	errServerHost = errors.New("serverHost environment variable is missing")
	errHTTPPort   = errors.New("httpPort environment variable is missing")
	errTCPPort    = errors.New("tcpPort environment variable is missing")
)

type Config struct {
	DatabaseURL    string
	AppPort        string
	AppHost        string
	HTTPPort       string
	TCPPort        string
	SigningKey     []byte
	ServerHost     string
	TCPServerAddr  string
	HTTPServerAddr string
	HTTPServerURL  string
	LoginURL       string
	RegisterURL    string
}

func New(log *logrus.Logger, path string) (*Config, error) {
	err := godotenv.Load(path)
	if err != nil {
		log.WithError(err).Panic("Error loading .env file")
	}

	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	postgresDB := os.Getenv("POSTGRES_DB")
	appHost := os.Getenv("APP_HOST")
	appPort := os.Getenv("APP_PORT")
	tcpPort := os.Getenv("TCP_PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	signingKey := []byte(jwtSecret)

	var dsn string

	switch {
	case postgresHost == "":
		return nil, errMissingHost
	case postgresPort == "":
		return nil, errMissingPort
	case postgresUser == "":
		return nil, errMissingUser
	case postgresPassword == "":
		return nil, errMissingPassword
	case postgresDB == "":
		return nil, errMissingDB
	case appPort == "":
		return nil, errMissingAppPort
	case tcpPort == "":
		return nil, errMissingTCPPort
	case jwtSecret == "" || len(signingKey) == 0:
		return nil, errMissingJwtSecret
	default:
		hostPort := net.JoinHostPort(postgresHost, postgresPort)
		dsn = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
			postgresUser, postgresPassword, hostPort, postgresDB)

		return &Config{
			DatabaseURL: dsn,
			AppPort:     appPort,
			AppHost:     appHost,
			TCPPort:     tcpPort,
			SigningKey:  signingKey,
		}, nil
	}
}

func NewClientConfig(log *logrus.Logger, path string) (*Config, error) {
	err := godotenv.Load(path)
	if err != nil {
		log.WithError(err).Panic("Error loading .env file")
	}

	serverHost := os.Getenv("SERVER_HOST")
	httpPort := os.Getenv("HTTP_PORT")
	tcpPort := os.Getenv("TCP_PORT")

	switch {
	case serverHost == "":
		return nil, errServerHost
	case httpPort == "":
		return nil, errHTTPPort
	case tcpPort == "":
		return nil, errTCPPort
	}

	const (
		userPath         = "/api/v1/user"
		loginEndpoint    = "/login"
		registerEndpoint = "/register"
	)

	tcpServerAddr := serverHost + ":" + tcpPort
	httpServerAddr := serverHost + ":" + httpPort
	httpServerURL := "http://" + httpServerAddr + userPath
	loginURL := httpServerURL + loginEndpoint
	registerURL := httpServerURL + registerEndpoint

	return &Config{
		ServerHost:     serverHost,
		HTTPPort:       httpPort,
		TCPPort:        tcpPort,
		TCPServerAddr:  tcpServerAddr,
		HTTPServerAddr: httpServerAddr,
		HTTPServerURL:  httpServerURL,
		LoginURL:       loginURL,
		RegisterURL:    registerURL,
	}, nil
}
