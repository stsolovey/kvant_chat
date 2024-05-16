package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/stsolovey/kvant_chat/internal/models"
)

type MockAuthRepo struct {
	mock.Mock
}

func (m *MockAuthRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*models.User), args.Error(1)
}

func setupAuthService() (*AuthService, *MockAuthRepo) {
	mockRepo := new(MockAuthRepo)
	signingKey := []byte("your-256-bit-secret")
	authService := NewAuthService(mockRepo, signingKey)
	return authService, mockRepo
}

func TestGenerateToken(t *testing.T) {
	authService, _ := setupAuthService()
	username := "testuser"

	token, err := authService.GenerateToken(username)
	assert.Nil(t, err, "should not error out when generating a token")

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-256-bit-secret"), nil
	})
	assert.Nil(t, err, "should be able to parse the token")
	assert.True(t, parsedToken.Valid, "token should be valid")

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok, "claims should be of type jwt.MapClaims")
	assert.Equal(t, username, claims["username"], "username should match")
	assert.True(t, claims["exp"].(float64) > float64(time.Now().Unix()), "token expiry should be in the future")
}

func TestValidateToken(t *testing.T) {
	authService, _ := setupAuthService()
	username := "testuser"
	token, _ := authService.GenerateToken(username)

	validatedToken, err := authService.ValidateToken(token)
	assert.Nil(t, err, "token validation should succeed")
	assert.NotNil(t, validatedToken, "validated token should not be nil")
}

func TestVerifyPassword(t *testing.T) {
	authService, _ := setupAuthService()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	result := authService.VerifyPassword(string(hashedPassword), "password123")
	assert.True(t, result, "password verification should succeed with correct password")

	result = authService.VerifyPassword(string(hashedPassword), "wrongpassword")
	assert.False(t, result, "password verification should fail with incorrect password")
}

func TestGetUserByUsername(t *testing.T) {
	authService, mockRepo := setupAuthService()
	ctx := context.Background()
	expectedUser := &models.User{UserName: "testuser", HashPassword: "hashed"}

	mockRepo.On("GetUserByUsername", ctx, "testuser").Return(expectedUser, nil)

	user, err := authService.GetUserByUsername(ctx, "testuser")
	assert.Nil(t, err, "should not error out when fetching user")
	assert.Equal(t, expectedUser, user, "returned user should match expected")
	mockRepo.AssertExpectations(t)
}
