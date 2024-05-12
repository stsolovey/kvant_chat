package httpserver

import (
	"github.com/go-chi/chi"
	"github.com/sirupsen/logrus"
	"github.com/stsolovey/kvant_chat/internal/app/handler"
	"github.com/stsolovey/kvant_chat/internal/app/service"
)

func configureRoutes(
	r chi.Router,
	log *logrus.Logger,
	usersServ service.UsersServiceInterface,
	authServ service.AuthServiceInterface,
) {
	authHandler := handler.NewAuthHandler(authServ, log)
	usersHandler := handler.NewUsersHandler(usersServ, log)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/login", authHandler.Login)            // .With(middleware.RateLimiterMiddleware).
			r.Post("/register", usersHandler.RegisterUser) // .With(middleware.RateLimiterMiddleware).
		})
	})
}

// r.Get("/", usersHandler.GetUsers).
// r.Route("/{id}", func(r chi.Router) {.
// 	r.Get("/", usersHandler.GetUser).
// 	r.Patch("/", usersHandler.UpdateUser).
//	r.Delete("/", usersHandler.DeleteUser).
// }).
