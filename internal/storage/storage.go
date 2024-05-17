package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // Importing `pgx/v5/stdlib` is necessary for `sql.Open("pgx", s.dsn)`.
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

var (
	ErrDBConnectionFailed  = errors.New("database connection failed")
	ErrTimeoutWaitingForDB = errors.New("timeout waiting for DB to be ready")
)

type Storage struct {
	db  *pgxpool.Pool
	dsn string
}

func (s *Storage) DB() *pgxpool.Pool {
	return s.db
}

func NewStorage(ctx context.Context, dsn string) (*Storage, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("storage.go NewStorage, pgxpool.ParseConfig(...): %w", err)
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("storage.go NewStorage, pgxpool.NewWithConfig(...): %w", err)
	}

	return &Storage{
		db:  db,
		dsn: dsn,
	}, nil
}

func (s *Storage) Migrate(logger *logrus.Logger) error {
	files, _ := migrationFiles.ReadDir("migrations")
	for _, file := range files {
		logger.Infof("Found migration file: %s", file.Name())
	}

	conn, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("storage.go Migrate, sql.Open(...) failed: %w", err)
	}
	defer conn.Close()

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationFiles,
		Root:       "migrations",
	}

	n, err := migrate.Exec(conn, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("migrate.Exec(...) failed: %w", err)
	}

	logger.Infof("Applied %d migrations successfully", n)

	return nil
}

func (s *Storage) WaitForDatabase(ctx context.Context, logger *logrus.Logger) error {
	const (
		tickerInterval  = 2 * time.Second
		timeoutDuration = 60 * time.Second
	)

	ticker := time.NewTicker(tickerInterval)
	timeout := time.After(timeoutDuration)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("storage.go WaitForDatabase context done: %w", ctx.Err())
		case <-timeout:
			return fmt.Errorf("storage.go WaitForDatabase timeout: %w", ErrTimeoutWaitingForDB)
		case <-ticker.C:
			if err := s.checkDB(); err == nil {
				logger.Info("Database is ready")

				return nil
			}

			logger.Info("Waiting for database to be ready")
		}
	}
}

func (s *Storage) checkDB() error {
	conn, err := sql.Open("pgx", s.dsn)
	if err != nil {
		return fmt.Errorf("storage.go checkDB, sql.Open failed: %w", ErrDBConnectionFailed)
	}
	defer conn.Close()

	if err = conn.Ping(); err != nil {
		return fmt.Errorf("storage.go checkDB, ping failed: %w", err)
	}

	return nil
}
