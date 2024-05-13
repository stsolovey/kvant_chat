package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stsolovey/kvant_chat/internal/models"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	args := m.Called(tokenString)
	var token *jwt.Token
	if args.Get(0) != nil {
		token = args.Get(0).(*jwt.Token)
	}
	return token, args.Error(1)
}

func (m *MockAuthService) GenerateToken(username string) (string, error) {
	return m.Called(username).Get(0).(string), m.Called(username).Error(1)
}

func (m *MockAuthService) VerifyPassword(storedHash, providedPassword string) bool {
	return m.Called(storedHash, providedPassword).Bool(0)
}

func (m *MockAuthService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return m.Called(ctx, username).Get(0).(*models.User), m.Called(ctx, username).Error(1)
}

func TestJWTAuthMiddleware(t *testing.T) {
	authService := new(MockAuthService)
	middleware := JWTAuthMiddleware(authService)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		token          string
		prepareMock    func()
		expectedStatus int
	}{
		{
			name:           "No token provided",
			token:          "",
			prepareMock:    func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "Invalid token",
			token: "invalidtoken",
			prepareMock: func() {
				authService.On("ValidateToken", "invalidtoken").Return(nil, jwt.ErrSignatureInvalid)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:  "Valid token",
			token: "validtoken",
			prepareMock: func() {
				claims := jwt.MapClaims{
					"username": "user123",
					"exp":      1700000000,
				}
				token := &jwt.Token{
					Valid:  true,
					Claims: claims,
				}
				authService.On("ValidateToken", "validtoken").Return(token, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.prepareMock()

			r := httptest.NewRequest("GET", "/api/v1/user", nil)
			if tc.token != "" {
				r.Header.Set("Authorization", "Bearer "+tc.token)
			}
			w := httptest.NewRecorder()

			handler := middleware(testHandler)
			handler.ServeHTTP(w, r)

			resp := w.Result()
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			authService.AssertExpectations(t)
		})
	}
}
