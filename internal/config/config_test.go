package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEnvironmentVariables(envVars map[string]string) {
	for key, value := range envVars {
		os.Setenv(key, value)
	}
}

func cleanupEnvironmentVariables(envVars []string) {
	for _, key := range envVars {
		os.Unsetenv(key)
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectError    error
		expectedConfig *Config
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"POSTGRES_HOST":     "localhost",
				"POSTGRES_PORT":     "5432",
				"POSTGRES_USER":     "user",
				"POSTGRES_PASSWORD": "password",
				"POSTGRES_DB":       "kvant_chat",
				"APP_HOST":          "127.0.0.1",
				"APP_PORT":          "8080",
				"JWT_SECRET":        "secret",
			},
			expectError: nil,
			expectedConfig: &Config{
				DatabaseURL: "postgres://user:password@localhost:5432/kvant_chat?sslmode=disable",
				AppPort:     "8080",
				AppHost:     "127.0.0.1",
				SigningKey:  []byte("secret"),
			},
		},
		{
			name:        "missing POSTGRES_HOST",
			envVars:     map[string]string{"POSTGRES_PORT": "5432"},
			expectError: errMissingHost,
		},
		{
			name:        "missing POSTGRES_PORT",
			envVars:     map[string]string{"POSTGRES_HOST": "localhost"},
			expectError: errMissingPort,
		},
		{
			name:        "missing POSTGRES_USER",
			envVars:     map[string]string{"POSTGRES_HOST": "localhost", "POSTGRES_PORT": "5432"},
			expectError: errMissingUser,
		},
		{
			name:        "missing POSTGRES_PASSWORD",
			envVars:     map[string]string{"POSTGRES_HOST": "localhost", "POSTGRES_PORT": "5432", "POSTGRES_USER": "user"},
			expectError: errMissingPassword,
		},
		{
			name:        "missing POSTGRES_DB",
			envVars:     map[string]string{"POSTGRES_HOST": "localhost", "POSTGRES_PORT": "5432", "POSTGRES_USER": "user", "POSTGRES_PASSWORD": "password"},
			expectError: errMissingDB,
		},
		{
			name:        "missing APP_PORT",
			envVars:     map[string]string{"POSTGRES_HOST": "localhost", "POSTGRES_PORT": "5432", "POSTGRES_USER": "user", "POSTGRES_PASSWORD": "password", "POSTGRES_DB": "kvant_chat"},
			expectError: errMissingAppPort,
		},
		{
			name:        "missing JWT_SECRET",
			envVars:     map[string]string{"POSTGRES_HOST": "localhost", "POSTGRES_PORT": "5432", "POSTGRES_USER": "user", "POSTGRES_PASSWORD": "password", "POSTGRES_DB": "kvant_chat", "APP_PORT": "8080"},
			expectError: errMissingJwtSecret,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setupEnvironmentVariables(tc.envVars)
			defer cleanupEnvironmentVariables([]string{"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB", "APP_HOST", "APP_PORT", "JWT_SECRET"})

			config, err := New()

			if tc.expectError != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectError, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedConfig, config)
			}
		})
	}
}
