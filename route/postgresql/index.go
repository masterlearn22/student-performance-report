package route

import (
    "database/sql"

    repo "student-performance-report/app/repository/postgresql"
    service "student-performance-report/app/service/postgresql"
    "student-performance-report/middleware"

    "github.com/gofiber/fiber/v2"
)

func SetupPostgresRoutes(app *fiber.App, db *sql.DB) {

    userRepo := repo.NewUserRepository(db)
    authService := service.NewAuthService(userRepo)

    adminRepo := repo.NewAdminRepository(db)
    adminService := service.NewAdminService(adminRepo, userRepo)

    // AUTH
    auth := app.Group("/api/v1/auth")
    auth.Post("/login", authService.Login)
    auth.Post("/refresh", authService.Refresh)
    auth.Post("/logout", middleware.AuthRequired(), authService.Logout)
    auth.Get("/profile", middleware.AuthRequired(), authService.Profile)

    // USERS (protected)
    users := app.Group("/api/v1/users", middleware.AuthRequired())
    users.Get("/", middleware.RoleAllowed("admin"), adminService.GetAllUsers)
    users.Get("/:id", adminService.GetUserByID)
    users.Post("/", middleware.RoleAllowed("admin"), adminService.CreateUser)
    users.Put("/:id", adminService.UpdateUser)
    users.Delete("/:id", adminService.DeleteUser)
    users.Put("/:id/role", middleware.RoleAllowed("admin"), adminService.AssignRole)
}
