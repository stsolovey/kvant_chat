package tests

import (
	"encoding/json"
	"net/http"

	"github.com/stsolovey/kvant_chat/internal/models"
)

func (s *IntegrationTestSuite) TestLogin() {
	s.Require().NoError(s.truncateTables())

	userInput := models.UserRegisterInput{
		UserName:     "TestUser",
		HashPassword: "TestPassword123.",
	}
	resp := s.sendRequest(s.ctx, http.MethodPost, pathRegister, userInput)
	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	s.Run("successful login", func() {
		loginInput := models.UserLoginInput{
			UserName: "TestUser",
			Password: "TestPassword123.",
		}

		resp := s.sendRequest(s.ctx, http.MethodPost, pathLogin, loginInput)
		defer resp.Body.Close()

		if !s.Assert().Equal(http.StatusOK, resp.StatusCode) {
			s.log.Errorf("Expected HTTP status %d, got %d for successful login", http.StatusOK, resp.StatusCode)
		}

		var response struct {
			Data struct {
				Token string `json:"token"`
			} `json:"data"`
		}
		err := json.NewDecoder(resp.Body).Decode(&response)
		if !s.Assert().NoError(err) {
			s.log.Errorf("Error decoding response for successful login: %v", err)
		}
		s.Assert().NotEmpty(response.Data.Token)
	})

	s.Run("login with incorrect password", func() {
		loginInput := models.UserLoginInput{
			UserName: "TestUser",
			Password: "IncorrectPassword",
		}

		resp := s.sendRequest(s.ctx, http.MethodPost, pathLogin, loginInput)
		defer resp.Body.Close()

		if !s.Assert().Equal(http.StatusUnauthorized, resp.StatusCode) {
			s.log.Errorf("Expected HTTP status %d for failed login due to incorrect password, got %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})

	s.Run("login with non-existent username", func() {
		loginInput := models.UserLoginInput{
			UserName: "NonExistentUser",
			Password: "AnyPassword",
		}

		resp := s.sendRequest(s.ctx, http.MethodPost, pathLogin, loginInput)
		defer resp.Body.Close()

		if !s.Assert().Equal(http.StatusUnauthorized, resp.StatusCode) {
			s.log.Errorf("Expected HTTP status %d for failed login due to non-existent username, got %d", http.StatusUnauthorized, resp.StatusCode)
		}
	})
}
