package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

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
	pathLogin    = "/api/v1/user/login"
	pathRegister = "/api/v1/user/register"
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
	s.cfg, err = config.New(s.log, "../.env")
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

	// Use a buffered channel to avoid blocking the goroutine
	errChan := make(chan error, 1)
	go func() {
		s.httpServer = httpserver.CreateServer(s.cfg, s.log, usersService, authService)
		errChan <- s.httpServer.Start(s.ctx)
	}()

	time.Sleep(1 * time.Second)
	select {
	case err := <-errChan:
		if err != nil {
			s.log.WithError(err).Error("Server start failed")
			s.T().FailNow()
		}
	default:
		s.log.Info("Server started successfully")
	}

	time.Sleep(100 * time.Millisecond)
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

	tables := []string{"users"}
	for _, table := range tables {
		_, err := s.storage.DB().Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			return fmt.Errorf("Failed to truncate table: %s; Error: %w", table, err)
		}
	}
	s.log.Infof("Tables truncated successfully")
	return nil
}
