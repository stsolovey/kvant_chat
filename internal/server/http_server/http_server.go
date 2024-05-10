package http_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/app/handler"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	"github.com/stsolovey/kvant_chat/internal/config"
)

type Server struct {
	config *config.Config
	logger *logrus.Logger
	server *http.Server
}

func configureRoutes(
	r chi.Router,
	log *logrus.Logger,
	usersServ service.UsersServiceInterface,
	authServ service.AuthServiceInterface,
) {
	authHandler := handler.NewAuthHandler(authServ, log)
	usersHandler := handler.NewUsersHandler(usersServ, log)

	r.Post("/login", authHandler.Login)

	r.Route("/api/v1/user", func(r chi.Router) {
		r.Post("/", usersHandler.CreateUser)
		r.Post("/register", usersHandler.RegisterUser)

		r.Get("/", usersHandler.GetUsers)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", usersHandler.GetUser)
			r.Patch("/", usersHandler.UpdateUser)
			r.Delete("/", usersHandler.DeleteUser)
		})
	})
}

func CreateServer(
	cfg *config.Config,
	log *logrus.Logger,
	port string,
	usersServ service.UsersServiceInterface,
	authServ service.AuthServiceInterface,
) *Server {
	r := chi.NewRouter()
	configureRoutes(r, log, usersServ, authServ)

	s := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
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

		ctxShutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
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
