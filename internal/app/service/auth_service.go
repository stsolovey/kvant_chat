package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stsolovey/kvant_chat/internal/app/repository"
	"github.com/stsolovey/kvant_chat/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInterface interface {
	GenerateToken(username string) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	VerifyPassword(storedHash, providedPassword string) bool
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
}

type AuthService struct {
	repo       repository.AuthRepositoryInterface
	signingKey []byte
}

func NewAuthService(repo repository.AuthRepositoryInterface, signingKey []byte) *AuthService {
	return &AuthService{
		repo:       repo,
		signingKey: signingKey,
	}
}

func (s *AuthService) GenerateToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", models.ErrTokenGenerationError
	}

	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString(s.signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, models.ErrTokenValidationError
		}

		return s.signingKey, nil
	})

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return token, nil
	}

	return nil, fmt.Errorf("validation error: %w", err)
}

func (s *AuthService) VerifyPassword(storedHash, providedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(providedPassword))

	return err == nil
}

func (s *AuthService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get user by username %s: %w", username, err)
	}

	return user, nil
}
