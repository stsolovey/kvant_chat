package tests

/*
import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stsolovey/kvant_chat/internal/models"
)

func (s *IntegrationTestSuite) TestRegister() {
	s.Require().NoError(s.truncateTables())

	s.Run("create multiple users", func() {
		for i := range 5 {
			userInput := models.UserRegisterInput{
				UserName:     fmt.Sprintf("TestUserName_%d", i+1),
				HashPassword: fmt.Sprintf("TestPassword_%d.", i+1),
			}

			resp := s.sendRequest(s.ctx, http.MethodPost, pathRegister, userInput)
			defer resp.Body.Close()

			if !s.Assert().Equal(http.StatusCreated, resp.StatusCode) {
				s.log.Errorf("Expected HTTP status %d, got %d for success registration %d", http.StatusCreated, resp.StatusCode, i+1)
			}

			var response struct {
				Data *models.UserRegisterInput `json:"data"`
			}
			err := json.NewDecoder(resp.Body).Decode(&response)
			if !s.Assert().NoError(err) {
				s.log.Errorf("Error decoding response for success registration %d: %v", i+1, err)
			}

			if !s.Assert().NotNil(response.Data) {
				s.log.Errorf("Expected success registration data to be non-nil for success registration %d", i+1)
			}
		}
	})

	s.Run("duplicate usernames", func() {
		userInput := models.UserRegisterInput{
			UserName:     "DuplicateUser",
			HashPassword: "ValidPassword123.",
		}

		// First registration should succeed
		resp := s.sendRequest(s.ctx, http.MethodPost, pathRegister, userInput)
		defer resp.Body.Close()
		s.Assert().Equal(http.StatusCreated, resp.StatusCode)

		// Second registration with the same username should fail
		resp = s.sendRequest(s.ctx, http.MethodPost, pathRegister, userInput)
		defer resp.Body.Close()

		if !s.Assert().Equal(http.StatusConflict, resp.StatusCode) {
			s.log.Errorf("Expected HTTP status %d for duplicate username, got %d", http.StatusConflict, resp.StatusCode)
		}
	})

	s.Run("invalid data formats", func() {
		testCases := []struct {
			desc   string
			user   models.UserRegisterInput
			status int
		}{
			{"empty username", models.UserRegisterInput{UserName: "", HashPassword: "password123"}, http.StatusBadRequest},
			{"empty password", models.UserRegisterInput{UserName: "user123", HashPassword: ""}, http.StatusBadRequest},
			{"short username", models.UserRegisterInput{UserName: "user", HashPassword: "password123"}, http.StatusBadRequest},
			{"short password", models.UserRegisterInput{UserName: "user123", HashPassword: "pass"}, http.StatusBadRequest},
		}

		for _, tc := range testCases {
			s.Run(tc.desc, func() {
				resp := s.sendRequest(s.ctx, http.MethodPost, pathRegister, tc.user)
				defer resp.Body.Close()
				if !s.Assert().Equal(tc.status, resp.StatusCode) {
					s.log.Errorf("Test '%s': expected HTTP status %d, got %d", tc.desc, tc.status, resp.StatusCode)
				}
			})
		}
	})
}
*/
