package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/stsolovey/kvant_chat/internal/app/repository"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	"github.com/stsolovey/kvant_chat/internal/config"
	"github.com/stsolovey/kvant_chat/internal/logger"
	httpserver "github.com/stsolovey/kvant_chat/internal/server/http-server"
	tcpserver "github.com/stsolovey/kvant_chat/internal/server/tcp-server"
	"github.com/stsolovey/kvant_chat/internal/storage"
	"golang.org/x/sync/errgroup"
)

func main() {
	log := logger.New()

	cfg, err := config.New(log, "./.env")
	if err != nil {
		log.WithError(err).Panic("Failed to initialize config")
	}

	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		log.Infof("Received signal: %s", sig)
		cancel()
	}()

	defer cancel()

	storageSystem, err := storage.NewStorage(ctx, cfg.DatabaseURL)
	if err != nil {
		log.WithError(err).Panic("Failed to initialize storage")
	}

	if err := storageSystem.Migrate(log); err != nil {
		log.WithError(err).Panic("Failed to execute migrations")
	}

	authRepo := repository.NewAuthRepository(storageSystem.DB())
	usersRepo := repository.NewUsersRepository(storageSystem.DB())

	authService := service.NewAuthService(authRepo, cfg.SigningKey)
	usersService := service.NewUsersService(usersRepo, authService)

	httpServer := httpserver.CreateServer(cfg, log, usersService, authService)
	tcpServer := tcpserver.CreateServer(cfg, log, authService)

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := tcpServer.Start(ctx); err != nil {
			log.WithError(err).Panic("TCP Server stopped unexpectedly")
		}

		return nil
	})

	eg.Go(func() error {
		if err := httpServer.Start(ctx); err != nil {
			log.WithError(err).Panic("Server stopped unexpectedly")
		}

		return nil
	})

	if err = eg.Wait(); err != nil {
		log.WithError(err).Panic("Server stopped unexpectedly")
	}
}
