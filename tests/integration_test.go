package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/stsolovey/kvant_chat/internal/app/repository"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	"github.com/stsolovey/kvant_chat/internal/config"
	"github.com/stsolovey/kvant_chat/internal/logger"
	httpserver "github.com/stsolovey/kvant_chat/internal/server/http-server"
	"github.com/stsolovey/kvant_chat/internal/storage"
)

const (
	pathLogin    = "localhost:8080/api/v1/user/login"
	pathRegister = "localhost:8080/api/v1/user/register"
)

type IntegrationTestSuite struct {
	suite.Suite
	log        *logrus.Logger
	httpServer *httpserver.Server
	storage    *storage.Storage
	ctx        context.Context
	cfg        *config.Config
	cancel     context.CancelFunc
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.log = logger.New()
	s.log.SetLevel(logrus.InfoLevel)
	s.ctx, s.cancel = context.WithCancel(context.Background())

	var err error
	if err := godotenv.Load("../.env"); err != nil {
		s.log.WithError(err).Error("Error loading .env file")
		s.T().FailNow()
	}

	s.cfg, err = config.New(s.log)
	if err != nil {
		s.log.WithError(err).Error("Failed to load configuration")
		s.T().FailNow()
	}

	s.storage, err = storage.NewStorage(s.ctx, s.cfg.DatabaseURL)
	if err != nil {
		s.log.WithError(err).Error("Failed to initialize storage")
		s.T().FailNow()
	}

	if err := s.storage.Migrate(s.log); err != nil {
		s.log.WithError(err).Error("Failed to execute migrations")
		s.T().FailNow()
	}

	authRepo := repository.NewAuthRepository(s.storage.DB())
	usersRepo := repository.NewUsersRepository(s.storage.DB())

	authService := service.NewAuthService(authRepo, s.cfg.SigningKey)
	usersService := service.NewUsersService(usersRepo, authService)

	go func() {
		s.httpServer = httpserver.CreateServer(s.cfg, s.log, usersService, authService)

		if err := s.httpServer.Start(s.ctx); err != nil {
			s.log.WithError(err).Error("Server start failed")
			s.T().FailNow()
		}
	}()

	time.Sleep(100 * time.Millisecond) // может быть нужно увеличить время
}

func (s *IntegrationTestSuite) TearDownSuite() {
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(s.ctx); err != nil {
			s.log.WithError(err).Error("Server shutdown failed")
		}
	}
	s.cancel()              // Отмена контекста
	time.Sleep(time.Second) // Ждём завершения
}

func (s *IntegrationTestSuite) sendRequest(
	ctx context.Context,
	method string,
	endpoint string,
	body any,
) *http.Response {
	s.T().Helper()

	reqBody, err := json.Marshal(body)
	s.Require().NoError(err)
	// fmt.Printf("Host: %s, Port: %s\n", s.cfg.AppHost, s.cfg.AppPort)
	req, err := http.NewRequestWithContext(ctx,
		method,
		fmt.Sprintf("http://%s:%s%s",
			s.cfg.AppHost,
			s.cfg.AppPort,
			endpoint), bytes.NewReader(reqBody))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)

	return resp
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) truncateTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tables := []string{"programs", "training_days", "exercises"}
	for _, table := range tables {
		_, err := s.storage.DB().Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			return fmt.Errorf("Failed to truncate table: %s; Error: %w", table, err)
		}
	}
	s.log.Infof("Tables truncated successfully")
	return nil
}
