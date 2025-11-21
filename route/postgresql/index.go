package route

import (
	"database/sql"

	repo "student-performance-report/app/repository/postgresql"
	service "student-performance-report/app/service/postgresql"
	// "student-performance-report/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupPostgresRoutes(app *fiber.App, db *sql.DB) {

	userRepo := repo.NewUserRepository(db)
	authService := service.NewAuthService(userRepo)

	api := app.Group("/api/v1/auth")

	// PUBLIC
	api.Post("/login", authService.Login)
	api.Post("/refresh", authService.Refresh)

	// PROTECTED
	// protected := api.Group("", middleware.AuthRequired())
	// protected.Post("/logout", authService.Logout)
	// protected.Get("/profile", authService.Profile)
}
