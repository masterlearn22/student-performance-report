package route

import (
    "database/sql"
    "github.com/gofiber/fiber/v2"
    repoMongo "student-performance-report/app/repository/mongodb"
    repoPostgre "student-performance-report/app/repository/postgresql"
    mongoService "student-performance-report/app/service/mongodb"
    postgreService "student-performance-report/app/service/postgresql"
    "student-performance-report/database"
    "student-performance-report/middleware"
)

func SetupPostgresRoutes(app *fiber.App, db *sql.DB) {
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
    achievementService := mongoService.NewAchievementService(achRepoMongo, achRepoPg, lecturerRepo)
	reportService := mongoService.NewReportService(achRepoMongo, studentRepo)

    // Static Files Config
    app.Static("/uploads", "./uploads")   
    api := app.Group("/api/v1")

    // 5.1 Authentication
    auth := api.Group("/auth")
    auth.Post("/login", authService.Login)
    auth.Post("/refresh", authService.Refresh)
    auth.Post("/logout",
        middleware.AuthRequired(),
        authService.Logout)
    auth.Get("/profile",
        middleware.AuthRequired(),  
        authService.Profile)

    // 5.2 Users 
    users := api.Group("/users", 
        middleware.AuthRequired(), 
        middleware.RoleAllowed("admin"))

    users.Get("/", adminService.GetAllUsers)
    users.Get("/:id", adminService.GetUserByID)
    users.Post("/", adminService.CreateUser)
    users.Put("/:id", adminService.UpdateUser)
    users.Delete("/:id", adminService.DeleteUser)
    users.Put("/:id/role", adminService.AssignRole)

    // 5.4 Achievements
    ach := api.Group("/achievements", middleware.AuthRequired())
    ach.Get("/", 
        middleware.PermissionRequired("achievement:read"), 
        achievementService.GetAllAchievements)
    
    ach.Get("/:id", 
        middleware.PermissionRequired("achievement:read"), 
        achievementService.GetAchievementDetail)
    
    ach.Get("/:id/history", 
        middleware.PermissionRequired("achievement:read"), 
        achievementService.GetAchievementHistory)

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

    // 5.5 Students & Lecturers
    api.Get("/students", 
        middleware.AuthRequired(),
        studentService.GetAllStudents)

    api.Get("/students/:id", 
        middleware.AuthRequired(), 
        studentService.GetStudentByID)

    api.Get("/students/:id/achievements", 
        middleware.AuthRequired(), 
        studentService.GetStudentAchievements)

    api.Put("/students/:id/advisor", 
        middleware.AuthRequired(), 
        studentService.UpdateAdvisor)

    api.Get("/lecturers", 
        middleware.AuthRequired(), 
        lecturerService.GetAllLecturers)

    api.Get("/lecturers/:id/advisees", 
        middleware.AuthRequired(), 
        lecturerService.GetAdvisees)

	// 5.8 Reports & Analytics (NEW)
	reports := api.Group("/reports", 
    middleware.AuthRequired())
        middleware.AuthRequired()
        
	reports.Get("/statistics",
		middleware.RoleAllowed("admin"),
		reportService.GetStatistics)

	reports.Get("/student/:id",
		reportService.GetStudentReport)
}