package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/stsolovey/kvant_chat/internal/app/repository"
	"github.com/stsolovey/kvant_chat/internal/models"
	"github.com/stsolovey/kvant_chat/internal/utils"
)

type UsersServiceInterface interface {
	RegisterUser(ctx context.Context, input models.UserRegisterInput) (*models.UserResponse, string, error)
}

type UsersService struct {
	repo        repository.UsersRepositoryInterface
	authService AuthServiceInterface
}

func NewUsersService(
	repo repository.UsersRepositoryInterface,
	authService AuthServiceInterface,
) UsersServiceInterface {
	return &UsersService{
		repo:        repo,
		authService: authService,
	}
}

func (s *UsersService) RegisterUser(ctx context.Context, input models.UserRegisterInput) (*models.UserResponse, string, error) {
	if len(input.UserName) < 6 {
		return nil, "", models.ErrUsernameTooShort
	}

	if len(input.HashPassword) < 6 {
		return nil, "", models.ErrPasswordTooShort
	}

	_, err := s.repo.GetUserByUsername(ctx, input.UserName)

	switch {
	case err == nil:
		return nil, "", models.ErrUsernameExists
	case !errors.Is(err, models.ErrUserNotFound):
		return nil, "", fmt.Errorf("users service RegisterUser(..) GetUserByUsername(...) error: %w", err)
	}

	hashedPassword, err := utils.HashPassword(input.HashPassword)
	if err != nil {
		return nil, "", fmt.Errorf("users service RegisterUser(..) utils.HashPassword(...) error: %w", err)
	}

	input.HashPassword = hashedPassword

	user, err := s.repo.Create(ctx, models.User{
		UserName:     input.UserName,
		HashPassword: input.HashPassword,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	userResponse := &models.UserResponse{
		ID:        user.ID,
		UserName:  user.UserName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	token, err := s.authService.GenerateToken(user.UserName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return userResponse, token, nil
}
