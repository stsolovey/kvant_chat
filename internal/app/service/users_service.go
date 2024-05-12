package service

import (
	"context"
	"fmt"

	"github.com/stsolovey/kvant_chat/internal/app/repository"
	"github.com/stsolovey/kvant_chat/internal/models"
	"github.com/stsolovey/kvant_chat/internal/util"
)

type UsersServiceInterface interface {
	RegisterUser(ctx context.Context, input models.UserRegisterInput) (*models.User, string, error)
	GetUser(ctx context.Context, id int) (*models.User, error)
	GetUsers(ctx context.Context, req models.FeedUsersRequest) ([]models.User, error)
	UpdateUser(ctx context.Context, id int, input models.UserUpdateInput) (*models.User, error)
	DeleteUser(ctx context.Context, id int) error
}

type UsersService struct {
	repo        repository.UsersRepositoryInterface
	authService AuthServiceInterface
}

func NewUsersService(repo repository.UsersRepositoryInterface) UsersServiceInterface {
	return &UsersService{repo: repo}
}

func (s *UsersService) RegisterUser(ctx context.Context, input models.UserRegisterInput) (*models.User, string, error) {
	if len(input.UserName) < 6 {
		return nil, "", models.ErrUsernameTooShort
	}
	if len(input.HashPassword) < 6 {
		return nil, "", models.ErrPasswordTooShort
	}

	_, err := s.repo.GetUserByUsername(ctx, input.UserName)
	if err == nil {
		return nil, "", models.ErrUsernameExists
	}

	hashedPassword, err := util.HashPassword(input.HashPassword)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}
	input.HashPassword = string(hashedPassword)

	user, err := s.repo.Create(ctx, models.User{
		UserName:     input.UserName,
		HashPassword: input.HashPassword,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}
	token, err := s.authService.GenerateToken(user.UserName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}
	return user, token, nil
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
	userToUpdate, _ := s.GetUser(ctx, id)

	updatedUser, _ := s.repo.Update(ctx, id, *userToUpdate)

	return updatedUser, nil
}

func (s *UsersService) DeleteUser(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("service: failed to delete user by id %d: %w", id, err)
	}

	return nil
}
