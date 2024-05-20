package httpserver

import (
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/app/handler"
	"github.com/stsolovey/kvant_chat/internal/app/service"
	"github.com/stsolovey/kvant_chat/internal/middleware"
	"golang.org/x/time/rate"
)

func configureRoutes(
	r chi.Router,
	log *logrus.Logger,
	usersServ service.UsersServiceInterface,
	authServ service.AuthServiceInterface,
) {
	authHandler := handler.NewAuthHandler(authServ, log)
	usersHandler := handler.NewUsersHandler(usersServ, log)

	const (
		loginRequestsPerSecond = 5
		loginBurstSize         = 15
		regisRequestsPerSecond = 5
		regisBurstSize         = 15
	)

	loginLimiter := rate.NewLimiter(loginRequestsPerSecond, loginBurstSize)
	regisLimiter := rate.NewLimiter(regisRequestsPerSecond, regisBurstSize)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.With(middleware.RateLimiterMiddleware(loginLimiter)).Post("/login", authHandler.Login)
			r.With(middleware.RateLimiterMiddleware(regisLimiter)).Post("/register", usersHandler.RegisterUser)
		})
	})
}
