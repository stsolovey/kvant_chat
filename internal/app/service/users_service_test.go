package service

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stsolovey/kvant_chat/internal/models"
)

// Mock UsersRepositoryInterface
type MockUsersRepo struct {
	mock.Mock
}

func (m *MockUsersRepo) Create(ctx context.Context, user models.User) (*models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUsersRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) != nil {
		return args.Get(0).(*models.User), args.Error(1)
	}
	return nil, args.Error(1)
}

// Mock AuthServiceInterface including all methods
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GenerateToken(username string) (string, error) {
	args := m.Called(username)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	args := m.Called(tokenString)
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockAuthService) VerifyPassword(storedHash, providedPassword string) bool {
	args := m.Called(storedHash, providedPassword)
	return args.Bool(0)
}

func (m *MockAuthService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*models.User), args.Error(1)
}

func setupUsersService() (*UsersService, *MockUsersRepo, *MockAuthService) {
	mockUsersRepo := new(MockUsersRepo)
	mockAuthService := new(MockAuthService)
	usersService := NewUsersService(mockUsersRepo, mockAuthService).(*UsersService) // Type assertion added
	return usersService, mockUsersRepo, mockAuthService
}

func TestRegisterUser(t *testing.T) {
	usersService, mockUsersRepo, mockAuthService := setupUsersService()
	ctx := context.Background()
	input := models.UserRegisterInput{
		UserName:     "newuser",
		HashPassword: "password123",
	}

	// Setup expected responses
	expectedUser := &models.User{
		UserName:     "newuser",
		HashPassword: "hashedpassword123",
	}
	expectedToken := "token123"
	expectedResponse := &models.UserResponse{
		UserName: "newuser",
	}

	// user not found (to allow creation) expected
	mockUsersRepo.On("GetUserByUsername", ctx, "newuser").Return(nil, models.ErrUserNotFound).Once()
	mockUsersRepo.On("Create", ctx, mock.AnythingOfType("models.User")).Return(expectedUser, nil).Once()
	mockAuthService.On("GenerateToken", "newuser").Return(expectedToken, nil).Once()

	// register should succeed
	userResponse, token, err := usersService.RegisterUser(ctx, input)
	assert.NoError(t, err, "RegisterUser should not return an error on first attempt")
	assert.Equal(t, expectedResponse.UserName, userResponse.UserName, "Username should match expected username on first attempt")
	assert.Equal(t, expectedToken, token, "Token should match expected token on first attempt")

	// user found (should block creation and throw error) expected
	mockUsersRepo.On("GetUserByUsername", ctx, "newuser").Return(expectedUser, nil).Once()

	// attempt to register should fail with username exists error
	userResponse, token, err = usersService.RegisterUser(ctx, input)
	assert.Error(t, err, "Should return error when username already exists")
	assert.Equal(t, models.ErrUsernameExists, err, "Error should be 'username already exists'")
}
