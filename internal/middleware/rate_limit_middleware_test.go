package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimiterMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		requests       int
		sleep          time.Duration
		expectedStatus []int
	}{
		{
			name:           "Single request within limit",
			requests:       1,
			expectedStatus: []int{http.StatusOK},
		},
		{
			name:           "Multiple requests exceed limit",
			requests:       12, // Make more requests than the burst size to test limit exceeding
			expectedStatus: []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests, http.StatusTooManyRequests},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			limiter := rate.NewLimiter(10, 2)
			middleware := RateLimiterMiddleware(limiter, nextHandler)

			for i := 0; i < tc.requests; i++ {
				r := httptest.NewRequest("GET", "/", nil)
				w := httptest.NewRecorder()
				middleware.ServeHTTP(w, r)
				if i < len(tc.expectedStatus) {
					assert.Equal(t, tc.expectedStatus[i], w.Code, "Request %d did not meet expected status", i+1)
				}
				time.Sleep(tc.sleep) // Sleep to simulate time passing if necessary
			}
		})
	}
}
