package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stsolovey/kvant_chat/internal/models"
)

type UsersRepositoryInterface interface {
	Create(ctx context.Context, user models.User) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type UsersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(db *pgxpool.Pool) UsersRepositoryInterface {
	return &UsersRepository{db: db}
}

func (r *UsersRepository) Create(
	ctx context.Context,
	user models.User,
) (*models.User, error) {
	var createdUser models.User

	currentTimestamp := "NOW()" // trying to avoid linter issues (`NOW()` duplication).
	sql := `INSERT INTO users (username, hashed_password, created_at, updated_at, deleted)
    VALUES ($1, $2, ` + currentTimestamp + `, ` + currentTimestamp + `, false)
    RETURNING user_id, username, hashed_password, created_at, updated_at, deleted`

	err := r.db.QueryRow(
		ctx,
		sql,
		user.UserName,
		user.HashPassword,
	).Scan(
		&createdUser.ID,
		&createdUser.UserName,
		&createdUser.HashPassword,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
		&createdUser.Deleted,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return &createdUser, nil
}

func (r *UsersRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*models.User, error) {
	var user models.User

	sql := `SELECT user_id, username, hashed_password, created_at, updated_at, deleted 
	FROM users WHERE username = $1`

	err := r.db.QueryRow(ctx, sql, username).Scan(
		&user.ID,
		&user.UserName,
		&user.HashPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Deleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}

		return nil, fmt.Errorf("users repository GetUserByUsername: %w", err)
	}

	return &user, nil
}
