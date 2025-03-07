package main

import (
	"auth-service/internal/config"
	"auth-service/internal/controller"
	"auth-service/internal/logger"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"net/http"
)

func main() {
	logs := logger.NewLogger()

	cfg := config.LoadConfig(logs)

	db := connectToDB(cfg, logs)
	defer db.Close()

	jwtService := service.NewJWTService(cfg["JWT_SECRET"])
	userRepo := repository.NewUserRepository(db, logs)
	userService := service.NewUserService(userRepo, logs, jwtService)
	userHandler := controller.NewUserHandler(userService, logs)
	jwtMiddleware := middleware.NewJWTMiddleware(jwtService)

	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", userHandler.RegisterHandler)
			r.Post("/login", userHandler.LoginHandler)
			r.Post("/refresh", userHandler.RefreshTokenHandler)

			r.Group(func(protected chi.Router) {
				protected.Use(jwtMiddleware.Authenticate)
				protected.Get("/users/me", userHandler.GetCurrentUserHandler)
				protected.Put("/user/me/update", userHandler.UpdateCurrentUserHandler)
				protected.Delete("/user/me/delete", userHandler.DeleteCurrentUser)
			})
		})
	})

	logs.Info.Printf("Auth service running on port %s", ":8081")
	err := http.ListenAndServe(":8081", r)
	logs.Info.Fatalf("Can't start server: %v", err)
}

func connectToDB(config map[string]string, logs *logger.Logger) *sql.DB {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config["POSTGRES_USER"], config["POSTGRES_PASSWORD"], config["POSTGRES_HOST"], config["POSTGRES_PORT"], config["POSTGRES_DB"])
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logs.Error.Fatalf("could not connect to database: %v", err)
	}
	return db
}

//func setupRoutes(userHandler *handler.UserController, logs *logger.Logger, jwtMiddleware *middleware.JWTMiddleware) *chi.Mux {
//	r := chi.NewRouter()
//	r.Route("/api/v1/auth", func(r chi.Router) {
//		r.Post("/register", userHandler.RegisterHandler)
//		r.Post("/login", userHandler.LoginHandler)
//		r.Post("/refresh", userHandler.RefreshToken)
//	})
//
//	r.Group(func(protected chi.Router) {
//		protected.Use(jwtMiddleware.Authenticate)
//		protected.Get("/api/v1/users/me", userHandler.GetCurrentUserHandler)
//	})
//
//	logs.Info.Println("Routes initialized")
//	return r
//}
