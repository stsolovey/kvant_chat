package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stsolovey/kvant_chat/internal/models"
)

type AuthRepositoryInterface interface {
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type AuthRepository struct {
	db *pgxpool.Pool
}

func NewAuthRepository(db *pgxpool.Pool) AuthRepositoryInterface {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetUserByUsername(
	ctx context.Context,
	username string,
) (*models.User, error) {
	var user models.User

	sql := `SELECT user_id, username, hashed_password, created_at, updated_at, deleted 
	FROM users WHERE username = $1
	AND NOT deleted`

	err := r.db.QueryRow(ctx, sql, username).Scan(
		&user.ID,
		&user.UserName,
		&user.HashPassword,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Deleted,
	)
	if err != nil {
		return nil, fmt.Errorf("auth repository GetUserByUsername error: %w", err)
	}

	return &user, nil
}
