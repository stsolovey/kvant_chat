package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stsolovey/kvant_chat/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInterface interface {
	GenerateToken(username string) (string, error)
	VerifyPassword(storedHash, providedPassword string) bool
}

type AuthService struct {
	signingKey []byte
}

func NewAuthService(signingKey []byte) *AuthService {
	return &AuthService{
		signingKey: signingKey,
	}
}

func (s *AuthService) GenerateToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", models.ErrInvalidTokenClaims
	}

	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString(s.signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (s *AuthService) VerifyPassword(storedHash, providedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(providedPassword))

	return err == nil
}
