package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/stsolovey/kvant_chat/internal/app/repository"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	"github.com/stsolovey/kvant_chat/internal/config"
	"github.com/stsolovey/kvant_chat/internal/logger"
	"github.com/stsolovey/kvant_chat/internal/server/http_server"
	"github.com/stsolovey/kvant_chat/internal/storage"
)

func main() {
	log := logger.New()

	err := godotenv.Load()
	if err != nil {
		log.WithError(err).Panic("Error loading .env file")
	}

	cfg, err := config.New()
	if err != nil {
		log.WithError(err).Panic("Failed to initialize config")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	storageSystem, err := storage.NewStorage(ctx, cfg.DatabaseURL)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize storage")
	}

	if err := storageSystem.WaitForDatabase(ctx, log); err != nil {
		log.WithError(err).Panic("Failed to wait for database to be ready")
	}

	if err := storageSystem.Migrate(log); err != nil {
		log.WithError(err).Panic("Failed to execute migrations")
	}

	usersRepo := repository.NewUsersRepository(storageSystem.DB())

	authService := service.NewAuthService(cfg.SigningKey)
	usersService := service.NewUsersService(usersRepo)

	http_server := http_server.CreateServer(cfg, log, "8080", usersService, authService)

	if err := http_server.Start(ctx); err != nil {
		log.WithError(err).Panic("Server stopped unexpectedly")
	}

	<-ctx.Done()

	if err := http_server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Failed to shut down server gracefully")
	}
}
