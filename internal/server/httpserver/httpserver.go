package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	"github.com/stsolovey/kvant_chat/internal/config"
)

const (
	readHeaderTimeoutDuration = 10 * time.Second
	readTimeoutDuration       = 15 * time.Second
	writeTimeoutDuration      = 15 * time.Second
	idleTimeoutDuration       = 60 * time.Second

	shutdownTimeoutDuration = 5 * time.Second
)

type Server struct {
	config *config.Config
	logger *logrus.Logger
	server *http.Server
}

func CreateServer(
	cfg *config.Config,
	log *logrus.Logger,
	port string,
	usersServ service.UsersServiceInterface,
	authServ service.AuthServiceInterface,
) *Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	configureRoutes(r, log, usersServ, authServ)

	s := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: readHeaderTimeoutDuration,
		ReadTimeout:       readTimeoutDuration,
		WriteTimeout:      writeTimeoutDuration,
		IdleTimeout:       idleTimeoutDuration,
	}

	return &Server{
		config: cfg,
		logger: log,
		server: s,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting server...")

	go func() {
		<-ctx.Done()
		s.logger.Info("Server is shutting down...")

		ctxShutdown, cancel := context.WithTimeout(ctx, shutdownTimeoutDuration)
		defer cancel()

		if err := s.server.Shutdown(ctxShutdown); err != nil {
			s.logger.WithError(err).Error("Server shutdown failed")
		}
	}()

	s.logger.Infof("Server is running on %s", s.server.Addr)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
