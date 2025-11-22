package route

import (
    "database/sql"

    "github.com/gofiber/fiber/v2"

    // Import Repositories
    repoMongo "student-performance-report/app/repository/mongodb"
    repoPostgre "student-performance-report/app/repository/postgresql"

    // Import Services
    mongoService "student-performance-report/app/service/mongodb"
    postgreService "student-performance-report/app/service/postgresql"

    "student-performance-report/database"
    "student-performance-report/middleware"
)

func SetupPostgresRoutes(app *fiber.App, db *sql.DB) {
    // =================================================================
    // DEPENDENCY INJECTION
    // =================================================================

    // Repositories
    userRepo := repoPostgre.NewUserRepository(db)
    adminRepo := repoPostgre.NewAdminRepository(db)
    studentRepo := repoPostgre.NewStudentRepository(db)
    lecturerRepo := repoPostgre.NewLecturerRepository(db)

    achRepoPg := repoPostgre.NewAchievementRepoPostgres(db)
    achRepoMongo := repoMongo.NewAchievementRepository(database.MongoDB)

    // Services
    authService := postgreService.NewAuthService(userRepo)
    adminService := postgreService.NewAdminService(adminRepo, userRepo)
    lecturerService := postgreService.NewLecturerService(lecturerRepo)
    studentService := postgreService.NewStudentService(studentRepo, achRepoMongo)
    
    // Note: AchievementService inject 3 dependencies (Mongo, Postgres, Lecturer)
    achievementService := mongoService.NewAchievementService(achRepoMongo, achRepoPg, lecturerRepo)

    // Static Files Config
    app.Static("/uploads", "./uploads")

    // =================================================================
    // ROUTE DEFINITIONS
    // =================================================================
    
    api := app.Group("/api/v1")

    // ---------------------------------------------------------
    // 5.1 Authentication
    // ---------------------------------------------------------
    auth := api.Group("/auth")
    auth.Post("/login", authService.Login)
    auth.Post("/refresh", authService.Refresh)
    auth.Post("/logout", middleware.AuthRequired(), authService.Logout)
    auth.Get("/profile", middleware.AuthRequired(), authService.Profile)

    // ---------------------------------------------------------
    // 5.2 Users (Admin Only)
    // ---------------------------------------------------------
    users := api.Group("/users", middleware.AuthRequired(), middleware.RoleAllowed("admin"))
    users.Get("/", adminService.GetAllUsers)
    users.Get("/:id", adminService.GetUserByID)
    users.Post("/", adminService.CreateUser)
    users.Put("/:id", adminService.UpdateUser)
    users.Delete("/:id", adminService.DeleteUser)
    users.Put("/:id/role", adminService.AssignRole)

    // ---------------------------------------------------------
    // 5.4 Achievements
    // ---------------------------------------------------------
    ach := api.Group("/achievements", middleware.AuthRequired())

    // General Read (All Roles with Permission)
    ach.Get("/", 
        middleware.PermissionRequired("achievement:read"), 
        achievementService.GetAllAchievements)
    
    ach.Get("/:id", 
        middleware.PermissionRequired("achievement:read"), 
        achievementService.GetAchievementDetail)
    
    ach.Get("/:id/history", 
        middleware.PermissionRequired("achievement:read"), 
        achievementService.GetAchievementHistory)

    // Mahasiswa Operations
    mhsMiddleware := middleware.RoleAllowed("mahasiswa")
    
    ach.Post("/", 
        mhsMiddleware, 
        middleware.PermissionRequired("achievement:create"), 
        achievementService.CreateAchievement)
    
    ach.Put("/:id", 
        mhsMiddleware, 
        middleware.PermissionRequired("achievement:update"), 
        achievementService.UpdateAchievement)
    
    ach.Delete("/:id", 
        mhsMiddleware, 
        middleware.PermissionRequired("achievement:delete"), 
        achievementService.DeleteAchievement)
    
    ach.Post("/:id/submit", 
        mhsMiddleware, 
        middleware.PermissionRequired("achievement:update"), 
        achievementService.SubmitAchievement)
    
    ach.Post("/:id/attachments", 
        mhsMiddleware, 
        middleware.PermissionRequired("achievement:update"), 
        achievementService.UploadAttachments)

    // Dosen Wali Operations
    dosenMiddleware := middleware.RoleAllowed("dosen_wali")
    verifyPermission := middleware.PermissionRequired("achievement:verify")

    ach.Post("/:id/verify", 
        dosenMiddleware, 
        verifyPermission, 
        achievementService.VerifyAchievement)
    
    ach.Post("/:id/reject", 
        dosenMiddleware, 
        verifyPermission, 
        achievementService.RejectAchievement)

    // ---------------------------------------------------------
    // 5.5 Students & Lecturers
    // ---------------------------------------------------------
    // Students
    api.Get("/students", middleware.AuthRequired(), studentService.GetAllStudents)
    api.Get("/students/:id", middleware.AuthRequired(), studentService.GetStudentByID)
    api.Get("/students/:id/achievements", middleware.AuthRequired(), studentService.GetStudentAchievements)
    api.Put("/students/:id/advisor", middleware.AuthRequired(), studentService.UpdateAdvisor)

    // Lecturers
    api.Get("/lecturers", middleware.AuthRequired(), lecturerService.GetAllLecturers)
    api.Get("/lecturers/:id/advisees", middleware.AuthRequired(), lecturerService.GetAdvisees)
}