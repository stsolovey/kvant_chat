package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stsolovey/kvant_chat/internal/models"
)

type UsersRepositoryInterface interface {
	Create(ctx context.Context, user models.User) (*models.User, error)
	Get(ctx context.Context, id int) (*models.User, error)
	GetUsers(ctx context.Context, req models.FeedUsersRequest) ([]models.User, error)
	Update(ctx context.Context, id int, user models.User) (*models.User, error)
	Delete(ctx context.Context, id int) error
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
	sql := `
    INSERT INTO users (username, hashed_password, created_at, updated_at, deleted)
    VALUES ($1, $2, NOW(), NOW(), false)
    RETURNING user_id, username, hashed_password, created_at, updated_at, deleted
    `
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

func (r *UsersRepository) Get(
	ctx context.Context,
	id int,
) (*models.User, error) {
	var user models.User

	err := r.db.QueryRow(ctx,
		`SELECT user_id, name, description
		FROM users
		WHERE user_id = $1
		AND deleted NOT false`, id).Scan(
		&user.ID,
		&user.UserName,
		&user.HashPassword,
	)
	if err != nil {
		return nil, models.ErrUserNotFound
	}

	return &user, nil
}

func (r *UsersRepository) GetUsers(
	ctx context.Context,
	req models.FeedUsersRequest,
) ([]models.User, error) {
	var users []models.User

	var queryBuilder strings.Builder

	var queryParams []interface{}

	queryBuilder.WriteString("SELECT user_id, username, description FROM users")

	if req.Sorting == "" {
		req.Sorting = "user_id"
	}

	if req.Text != "" {
		queryParams = append(queryParams, "%"+req.Text+"%")
		queryBuilder.WriteString(fmt.Sprintf(" WHERE username LIKE $%d OR description LIKE $%d",
			len(queryParams), len(queryParams)))
	}

	validColumns := map[string]struct{}{"username": {}, "description": {}, "user_id": {}}
	if _, ok := validColumns[req.Sorting]; !ok {
		return nil, models.ErrInvalidSortingColumn
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY \"%s\"", req.Sorting))

	if req.Descending {
		queryBuilder.WriteString(" DESC")
	} else {
		queryBuilder.WriteString(" ASC")
	}

	queryParams = append(queryParams, req.Offset)
	queryParams = append(queryParams, req.ItemsPerPage)
	queryBuilder.WriteString(fmt.Sprintf(" OFFSET $%d LIMIT $%d", len(queryParams)-1, len(queryParams)))

	query := queryBuilder.String()

	rows, err := r.db.Query(ctx, query, queryParams...)
	if err != nil {
		return nil, fmt.Errorf("error querying users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p models.User
		if err := rows.Scan(&p.ID, &p.UserName, &p.HashPassword); err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}

		users = append(users, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *UsersRepository) Update(ctx context.Context, id int, user models.User) (*models.User, error) {
	var updatedUser models.User

	err := r.db.QueryRow(ctx,
		`UPDATE users SET username = $1, description = $2, updated_at=NOW()
		WHERE user_id = $3
        RETURNING user_id, username, description, created_at, updated_at, deleted`,
		user.UserName, user.HashPassword, id).Scan(
		&updatedUser.ID,
		&updatedUser.UserName,
		&updatedUser.HashPassword,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
		&updatedUser.Deleted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}

		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return &updatedUser, nil
}

func (r *UsersRepository) Delete(ctx context.Context, id int) error {
	var deletedFlag bool

	err := r.db.QueryRow(ctx,
		`UPDATE users SET deleted = true, updated_at = NOW()
		WHERE user_id = $1 AND deleted = false
		RETURNING deleted`,
		id).Scan(&deletedFlag)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ErrUserNotFound
		}

		return fmt.Errorf("error deleting user: %w", err)
	}

	if !deletedFlag {
		return models.ErrUserWasNotDeleted
	}

	return nil
}
