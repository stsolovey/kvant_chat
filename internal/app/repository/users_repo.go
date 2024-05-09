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

	err := r.db.QueryRow(ctx, `INSERT INTO users (..) VALUES ($..)
        RETURNING ..`,
		user.Name, user.HashPassword).Scan(
		&createdUser.ID,
		&createdUser.Name,
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
		WHERE user_id = $1`, id).Scan(
		&user.ID,
		&user.Name,
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

	queryBuilder.WriteString("SELECT user_id, name, description FROM users")

	if req.Sorting == "" {
		req.Sorting = "user_id"
	}

	if req.Text != "" {
		queryParams = append(queryParams, "%"+req.Text+"%")
		queryBuilder.WriteString(fmt.Sprintf(" WHERE name LIKE $%d OR description LIKE $%d",
			len(queryParams), len(queryParams)))
	}

	validColumns := map[string]struct{}{"name": {}, "description": {}, "user_id": {}}
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
		if err := rows.Scan(&p.ID, &p.Name, &p.HashPassword); err != nil {
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
		`UPDATE users SET name = $1, description = $2, updated_at=NOW()
		WHERE user_id = $3
        RETURNING user_id, name, description, created_at, updated_at, deleted`,
		user.Name, user.HashPassword, id).Scan(
		&updatedUser.ID,
		&updatedUser.Name,
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
