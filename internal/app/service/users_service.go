package service

import (
	"context"
	"fmt"

	"github.com/stsolovey/kvant_chat/internal/app/repository"
	"github.com/stsolovey/kvant_chat/internal/models"
)

type UsersServiceInterface interface {
	CreateUser(ctx context.Context, input models.UserCreateInput) (*models.User, error)
	GetUser(ctx context.Context, id int) (*models.User, error)
	GetUsers(ctx context.Context, req models.FeedUsersRequest) ([]models.User, error)
	UpdateUser(ctx context.Context, id int, input models.UserUpdateInput) (*models.User, error)
	DeleteUser(ctx context.Context, id int) error
}

type UsersService struct {
	repo repository.UsersRepositoryInterface
}

func NewUsersService(repo repository.UsersRepositoryInterface) UsersServiceInterface {
	return &UsersService{repo: repo}
}

func (s *UsersService) CreateUser(
	ctx context.Context,
	input models.UserCreateInput,
) (*models.User, error) {
	if input.Name == "" {
		return nil, models.ErrUserNameRequired
	} else if len(input.Name) < 3 {
		return nil, models.ErrUserNameTooShort
	}

	user := models.User{
		Name:         input.Name,
		HashPassword: input.HashPassword,
	}

	createdUser, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (s *UsersService) GetUser(ctx context.Context, id int) (*models.User, error) {
	user, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get user by id %d: %w", id, err)
	}

	return user, nil
}

func (s *UsersService) GetUsers(ctx context.Context, req models.FeedUsersRequest) ([]models.User, error) {
	if req.ItemsPerPage <= 0 {
		req.ItemsPerPage = 10 // Значение по умолчанию, если не задано
	}

	if req.Offset < 0 {
		req.Offset = 0 // Значение по умолчанию, если не задано
	}

	if req.ItemsPerPage > 100 { // ограничиваем максимум
		return nil, fmt.Errorf("UsersService GetUsers: %w", models.ErrValueExceededMaximum)
	}

	users, err := s.repo.GetUsers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get users: %w", err)
	}

	return users, nil
}

func (s *UsersService) UpdateUser(ctx context.Context, id int,
	input models.UserUpdateInput,
) (*models.User, error) {
	if input.Name == "" {
		return nil, models.ErrUserNameRequired
	} else if len(input.Name) < 3 {
		return nil, models.ErrUserNameTooShort
	}

	userToUpdate, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("id not found: %w", err)
	}

	if input.Name != "" {
		userToUpdate.Name = input.Name
	}

	if input.HashPassword != "" {
		userToUpdate.HashPassword = input.HashPassword
	}

	updatedUser, err := s.repo.Update(ctx, id, *userToUpdate)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return updatedUser, nil
}

func (s *UsersService) DeleteUser(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("service: failed to delete user by id %d: %w", id, err)
	}

	return nil
}
